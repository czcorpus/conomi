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

func (rdb *ReportsDatabase) selectActiveGroupID(sourceID general.SourceID) (int, error) {
	whereClause, whereValues := make([]string, 0, 4), make([]any, 0, 3)
	whereClause, whereValues = append(whereClause, "app = ?"), append(whereValues, sourceID.App)
	if sourceID.Instance != "" {
		whereClause, whereValues = append(whereClause, "instance = ?"), append(whereValues, sourceID.Instance)
	} else {
		whereClause = append(whereClause, "instance IS NULL")
	}
	if sourceID.Tag != "" {
		whereClause, whereValues = append(whereClause, "tag = ?"), append(whereValues, sourceID.Tag)
	} else {
		whereClause = append(whereClause, "tag IS NULL")
	}
	whereClause = append(whereClause, "resolved_by_user_id IS NULL")

	sql1 := "SELECT id FROM conomi_report_group " +
		"WHERE " + strings.Join(whereClause, " AND ") + " LIMIT 1"

	log.Debug().Str("sql", sql1).Msgf("going to SELECT conomi_report_group WHERE app = %s, instance = %s, tag = %s", sourceID.App, sourceID.Instance, sourceID.Tag)
	var groupID int
	row := rdb.db.QueryRow(sql1, whereValues...)
	if err := row.Scan(&groupID); err != nil {
		return 0, err
	}
	return groupID, nil
}

func (rdb *ReportsDatabase) createNewActiveGroup(sourceID general.SourceID) (int, error) {
	sql1 := "INSERT INTO conomi_report_group (app, instance, tag, severity) VALUES (?,?,?,?)"
	log.Debug().Str("sql", sql1).Msgf("going to INSERT conomi_report_group WHERE app = %s, instance = %s, tag = %s", sourceID.App, sourceID.Instance, sourceID.Tag)
	instance := sql.NullString{String: sourceID.Instance, Valid: sourceID.Instance != ""}
	tag := sql.NullString{String: sourceID.Tag, Valid: sourceID.Tag != ""}
	result, err := rdb.db.Exec(sql1, sourceID.App, instance, tag, general.SeverityLevelInfo)
	if err != nil {
		return -1, err
	}
	groupID, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(groupID), nil
}

func (rdb *ReportsDatabase) InsertReport(report general.Report) (int, int, error) {
	groupID, err := rdb.selectActiveGroupID(report.SourceID)
	if err == sql.ErrNoRows {
		groupID, err = rdb.createNewActiveGroup(report.SourceID)
		if err != nil {
			return -1, -1, err
		}
	}

	sql1 := "INSERT INTO conomi_report (report_group_id, severity, subject, body, args, created) VALUES (?,?,?,?,?,?)"
	entry, err := NewReportSQL(report)
	if err != nil {
		return -1, -1, err
	}
	log.Debug().Str("sql", sql1).Msg("going to INSERT report")
	result, err := rdb.db.Exec(sql1, groupID, entry.Severity, entry.Subject, entry.Body, entry.Args, entry.Created)
	if err != nil {
		return -1, -1, err
	}
	reportID, err := result.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	return int(reportID), groupID, nil
}

func (rdb *ReportsDatabase) ListReports(sourceID general.SourceID) ([]*general.Report, error) {
	whereClause := make([]string, 0, 4)
	whereValues := make([]any, 0, 3)
	whereClause = append(whereClause, "resolved_by_user_id IS NULL")
	if sourceID.App != "" {
		whereClause = append(whereClause, "app = ?")
		whereValues = append(whereValues, sourceID.App)
	}
	if sourceID.Instance != "" {
		whereClause = append(whereClause, "instance = ?")
		whereValues = append(whereValues, sourceID.Instance)
	}
	if sourceID.Tag != "" {
		whereClause = append(whereClause, "tag = ?")
		whereValues = append(whereValues, sourceID.Tag)
	}
	sql1 := "SELECT cr.id, crg.id, crg.app, crg.instance, crg.tag, cr.severity, cr.subject, cr.body, cr.args, cr.created, crg.resolved_by_user_id, us.user " +
		"FROM conomi_report_group AS crg " +
		"JOIN conomi_report AS cr ON crg.id = cr.report_group_id " +
		"LEFT JOIN user AS us ON resolved_by_user_id = us.id " +
		"WHERE " + strings.Join(whereClause, " AND ") + " " +
		"ORDER BY created DESC"
	log.Debug().Str("sql", sql1).Msg("going to SELECT conomi_report WHERE resolved_by_user_id IS NULL")
	rows, err := rdb.db.Query(sql1, whereValues...)
	if err != nil {
		return nil, err
	}
	ans := make([]*general.Report, 0, 100)
	for rows.Next() {
		entry := &ReportSQL{}
		err := rows.Scan(&entry.ID, &entry.GroupID, &entry.App, &entry.Instance, &entry.Tag, &entry.Severity, &entry.Subject, &entry.Body, &entry.Args, &entry.Created, &entry.ResolvedByUserID, &entry.ResolvedByUserName)
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
	sql1 := "SELECT cr.id, crg.id, crg.app, crg.instance, crg.tag, cr.severity, cr.subject, cr.body, cr.args, cr.created, crg.resolved_by_user_id, us.user " +
		"FROM conomi_report_group AS crg " +
		"JOIN conomi_report AS cr ON crg.id = cr.report_group_id " +
		"LEFT JOIN user AS us ON resolved_by_user_id = us.id " +
		"WHERE cr.id = ? LIMIT 1"
	log.Debug().Str("sql", sql1).Msgf("going to SELECT conomi_reports WHERE id = %d", reportID)
	entry := &ReportSQL{}
	row := rdb.db.QueryRow(sql1, reportID)
	if err := row.Scan(&entry.ID, &entry.GroupID, &entry.App, &entry.Instance, &entry.Tag, &entry.Severity, &entry.Subject, &entry.Body, &entry.Args, &entry.Created, &entry.ResolvedByUserID, &entry.ResolvedByUserName); err != nil {
		return nil, err
	}
	if err := entry.Severity.Validate(); err != nil {
		return nil, err
	}
	return entry.Export()
}

func (rdb *ReportsDatabase) ResolveGroup(groupID int, userID int) error {
	sql1 := "UPDATE conomi_report_group AS crg " +
		"SET crg.resolved_by_user_id = ? " +
		"WHERE crg.resolved_by_user_id IS NULL AND crg.id = ?"
	log.Debug().Str("sql", sql1).Msgf("going to resolve group WHERE id = %d", groupID)
	_, err := rdb.db.Exec(sql1, userID, groupID)
	if err != nil {
		return err
	}
	return nil
}

func (rdb *ReportsDatabase) GetReportCounts() ([]*general.ReportCount, error) {
	sql1 := "SELECT crg.app, crg.instance, crg.tag, " +
		"SUM(CASE WHEN cr.severity = ? THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN cr.severity = ? THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN cr.severity = ? THEN 1 ELSE 0 END) " +
		"FROM conomi_report_group AS crg " +
		"JOIN conomi_report AS cr ON crg.id = cr.report_group_id " +
		"WHERE crg.resolved_by_user_id IS NULL " +
		"GROUP BY crg.app, crg.instance, crg.tag ORDER BY crg.app, crg.instance, crg.tag"
	log.Debug().Str("sql", sql1).Msg("going to count conomi_report WHERE resolved_by_user_id IS NULL")
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
	sql1 := "SELECT DISTINCT app, instance, tag FROM conomi_report_group WHERE resolved_by_user_id IS NULL ORDER BY app, instance, tag"
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
