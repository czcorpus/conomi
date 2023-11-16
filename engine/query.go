// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package engine

import (
	"context"
	"database/sql"
	"strings"

	"github.com/czcorpus/conomi/general"
	"github.com/rs/zerolog/log"
)

// ReportsDatabase
// note: the lifecycle of the instance
// is "per request"
type ReportsDatabase struct {
	db  *sql.DB
	ctx context.Context
}

func (rdb *ReportsDatabase) InsertReport(report general.Report) (int, error) {
	sql1 := "INSERT INTO conomi_reports (app, instance, tag, severity, subject, body, args, created) VALUES (?,?,?,?,?,?,?,?)"
	entry, err := NewReportSQL(report)
	if err != nil {
		return -1, nil
	}
	log.Debug().Str("sql", sql1).Msg("going to INSERT report")
	result, err := rdb.db.Exec(sql1, entry.App, entry.Instance, entry.Tag, entry.Severity, entry.Subject, entry.Body, entry.Args, entry.Created)
	if err != nil {
		return -1, err
	}
	reportID, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(reportID), nil
}

func (rdb *ReportsDatabase) ListReports(app, instance, tag string) ([]*general.Report, error) {
	whereClause := make([]string, 0, 4)
	whereValues := make([]any, 0, 3)
	whereClause = append(whereClause, "resolved_by_user_id IS NULL")
	if app != "" {
		whereClause = append(whereClause, "app = ?")
		whereValues = append(whereValues, app)
	}
	if instance != "" {
		whereClause = append(whereClause, "instance = ?")
		whereValues = append(whereValues, instance)
	}
	if tag != "" {
		whereClause = append(whereClause, "tag = ?")
		whereValues = append(whereValues, tag)
	}
	sql1 := "SELECT id, app, instance, tag, severity, subject, body, args, created, resolved_by_user_id " +
		"FROM conomi_reports " +
		"WHERE " + strings.Join(whereClause, " AND ") + " " +
		"ORDER BY created DESC"
	log.Debug().Str("sql", sql1).Msg("going to SELECT conomi_reports WHERE resolved_by_user_id IS NULL")
	rows, err := rdb.db.Query(sql1, whereValues...)
	if err != nil {
		return nil, err
	}
	ans := make([]*general.Report, 0, 100)
	for rows.Next() {
		entry := &ReportSQL{}
		err := rows.Scan(&entry.ID, &entry.App, &entry.Instance, &entry.Tag, &entry.Severity, &entry.Subject, &entry.Body, &entry.Args, &entry.Created, &entry.ResolvedByUserID)
		if err != nil {
			return nil, err
		}
		if err := entry.Severity.Validate(); err != nil {
			return nil, err
		}
		item, err := entry.Export()
		if err != nil {
			return nil, err
		}
		ans = append(ans, item)
	}
	return ans, nil
}

func (rdb *ReportsDatabase) SelectReport(reportID int) (*general.Report, error) {
	sql1 := "SELECT cr.id, cr.app, cr.instance, cr.tag, cr.severity, cr.subject, cr.body, cr.args, cr.created, cr.resolved_by_user_id, us.user as resolved_by_user_name " +
		"FROM conomi_reports AS cr " +
		"LEFT JOIN users AS us " +
		"ON cr.resolved_by_user_id = us.id " +
		"WHERE cr.id = ? LIMIT 1"
	log.Debug().Str("sql", sql1).Msgf("going to SELECT conomi_reports WHERE id = %d", reportID)
	entry := &ReportSQL{}
	row := rdb.db.QueryRow(sql1, reportID)
	if err := row.Scan(&entry.ID, &entry.App, &entry.Instance, &entry.Tag, &entry.Severity, &entry.Subject, &entry.Body, &entry.Args, &entry.Created, &entry.ResolvedByUserID, &entry.ResolvedByUserName); err != nil {
		return nil, err
	}
	if err := entry.Severity.Validate(); err != nil {
		return nil, err
	}
	return entry.Export()
}

func (rdb *ReportsDatabase) ResolveReport(reportID int, userID int) (int, error) {
	sql1 := "UPDATE conomi_reports SET resolved_by_user_id = ? WHERE id = ? AND resolved_by_user_id IS NULL"
	log.Debug().Str("sql", sql1).Msgf("going to resolve report WHERE id = %d", reportID)
	result, err := rdb.db.Exec(sql1, userID, reportID)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	return int(rows), err
}

func (rdb *ReportsDatabase) ResolveGroup(reportID int, userID int) (int, error) {
	sql1 := "UPDATE conomi_reports AS upd " +
		"INNER JOIN conomi_reports AS sel " +
		"ON upd.app = sel.app AND upd.instance <=> sel.instance AND upd.tag <=> sel.tag AND sel.id = ? " +
		"SET upd.resolved_by_user_id = ? " +
		"WHERE upd.resolved_by_user_id IS NULL"
	log.Debug().Str("sql", sql1).Msgf("going to resolve all group reports WHERE id = %d", reportID)
	result, err := rdb.db.Exec(sql1, reportID, userID)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	return int(rows), err
}

func (rdb *ReportsDatabase) GetReportCounts() ([]*general.ReportCount, error) {
	sql1 := "SELECT app, instance, tag, " +
		"SUM(CASE WHEN severity = ? THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN severity = ? THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN severity = ? THEN 1 ELSE 0 END) " +
		"FROM conomi_reports WHERE resolved_by_user_id IS NULL " +
		"GROUP BY app, instance, tag ORDER BY app, instance, tag"
	log.Debug().Str("sql", sql1).Msg("going to count conomi_reports WHERE resolved_by_user_id IS NULL")
	rows, err := rdb.db.Query(sql1, general.SeverityLevelCritical, general.SeverityLevelWarning, general.SeverityLevelInfo)
	if err != nil {
		return nil, err
	}
	ans := make([]*general.ReportCount, 0, 100)
	for rows.Next() {
		count := &general.ReportCount{}
		var instance, tag sql.NullString
		err := rows.Scan(&count.SourceID.App, &instance, &tag, &count.Critical, &count.Warning, &count.Info)
		if err != nil {
			return nil, err
		}
		count.SourceID.Instance, count.SourceID.Tag = instance.String, tag.String
		ans = append(ans, count)
	}
	return ans, nil
}

func (rdb *ReportsDatabase) GetSources() ([]*general.SourceID, error) {
	sql1 := "SELECT DISTINCT app, instance, tag FROM conomi_reports WHERE resolved_by_user_id IS NULL ORDER BY app, instance, tag"
	log.Debug().Str("sql", sql1).Msg("going to get available filters")
	rows, err := rdb.db.Query(sql1)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	ans := make([]*general.SourceID, 0, 100)
	for rows.Next() {
		var app string
		var instance, tag sql.NullString
		err := rows.Scan(&app, &instance, &tag)
		if err != nil {
			return nil, err
		}
		ans = append(ans, &general.SourceID{
			App:      app,
			Instance: instance.String,
			Tag:      tag.String,
		})
	}
	return ans, nil
}

func NewReportsDatabase(db *sql.DB) *ReportsDatabase {
	return &ReportsDatabase{
		db:  db,
		ctx: context.Background(),
	}
}
