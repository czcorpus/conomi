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

func (rdb *ReportsDatabase) ListReports() ([]*general.Report, error) {
	sql1 := "SELECT id, app, instance, tag, severity, subject, body, args, created, resolved_by_user_id " +
		"FROM conomi_reports " +
		"WHERE resolved_by_user_id IS NULL"
	log.Debug().Str("sql", sql1).Msg("going to SELECT conomi_reports WHERE resolved_by_user_id IS NULL")
	rows, err := rdb.db.Query(sql1)
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
	sql1 := "SELECT id, app, instance, tag, severity, subject, body, args, created, resolved_by_user_id " +
		"FROM conomi_reports " +
		"WHERE id = ? LIMIT 1"
	log.Debug().Str("sql", sql1).Msgf("going to SELECT conomi_reports WHERE id = %d", reportID)
	entry := &ReportSQL{}
	row := rdb.db.QueryRow(sql1, reportID)
	if err := row.Scan(&entry.ID, &entry.App, &entry.Instance, &entry.Tag, &entry.Severity, &entry.Subject, &entry.Body, &entry.Args, &entry.Created, &entry.ResolvedByUserID); err != nil {
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

func (rdb *ReportsDatabase) ResolveReportsSince(reportID int, userID int) (int, error) {
	sql1 := "UPDATE conomi_reports AS upd " +
		"INNER JOIN conomi_reports AS sel " +
		"ON upd.app = sel.app AND upd.instance <=> sel.instance AND upd.tag <=> sel.tag AND upd.created >= sel.created AND sel.id = ? " +
		"SET upd.resolved_by_user_id = ? " +
		"WHERE upd.resolved_by_user_id IS NULL"
	log.Debug().Str("sql", sql1).Msgf("going to resolve new reports WHERE id = %d", reportID)
	result, err := rdb.db.Exec(sql1, reportID, userID)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	return int(rows), err
}

func (rdb *ReportsDatabase) GetReportCounts() ([]*general.ReportCount, error) {
	sql1 := "SELECT app, instance, tag, " +
		"SUM(CASE WHEN severity = \"critical\" THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN severity = \"warning\" THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN severity = \"info\" THEN 1 ELSE 0 END) " +
		"FROM conomi_reports WHERE resolved_by_user_id IS NULL GROUP BY app, instance, tag"
	log.Debug().Str("sql", sql1).Msg("going to count conomi_reports WHERE resolved_by_user_id IS NULL")
	rows, err := rdb.db.Query(sql1)
	if err != nil {
		return nil, err
	}
	ans := make([]*general.ReportCount, 0, 100)
	for rows.Next() {
		count := &general.ReportCount{}
		var instance, tag sql.NullString
		err := rows.Scan(&count.App, &instance, &tag, &count.Critical, &count.Warning, &count.Info)
		if err != nil {
			return nil, err
		}
		count.Instance, count.Tag = instance.String, tag.String
		ans = append(ans, count)
	}
	return ans, nil
}

func NewReportsDatabase(db *sql.DB) *ReportsDatabase {
	return &ReportsDatabase{
		db:  db,
		ctx: context.Background(),
	}
}
