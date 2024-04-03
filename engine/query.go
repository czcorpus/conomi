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
	"fmt"
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

func (rdb *ReportsDatabase) updateGroupID(report *general.Report) error {
	whereClause, whereValues := make([]string, 0, 4), make([]any, 0, 3)
	whereClause, whereValues = append(whereClause, "app = ?"), append(whereValues, report.SourceID.App)
	if report.SourceID.Instance != "" {
		whereClause, whereValues = append(whereClause, "instance = ?"), append(whereValues, report.SourceID.Instance)
	} else {
		whereClause = append(whereClause, "instance IS NULL")
	}
	if report.SourceID.Tag != "" {
		whereClause, whereValues = append(whereClause, "tag = ?"), append(whereValues, report.SourceID.Tag)
	} else {
		whereClause = append(whereClause, "tag IS NULL")
	}
	whereClause = append(whereClause, "resolved_by_user_id IS NULL")

	sql1 := "SELECT id FROM conomi_report_group " +
		"WHERE " + strings.Join(whereClause, " AND ") + " LIMIT 1"

	log.Debug().Str("sql", sql1).Msgf("going to SELECT conomi_report_group WHERE app = %s, instance = %s, tag = %s", report.SourceID.App, report.SourceID.Instance, report.SourceID.Tag)
	row := rdb.db.QueryRow(sql1, whereValues...)
	return row.Scan(&report.GroupID)
}

func (rdb *ReportsDatabase) assignNewGroup(report *general.Report) error {
	sql1 := "INSERT INTO conomi_report_group (app, instance, tag, created) VALUES (?,?,?,?)"
	log.Debug().Str("sql", sql1).Msgf("going to INSERT conomi_report_group WHERE app = %s, instance = %s, tag = %s", report.SourceID.App, report.SourceID.Instance, report.SourceID.Tag)
	instance := sql.NullString{String: report.SourceID.Instance, Valid: report.SourceID.Instance != ""}
	tag := sql.NullString{String: report.SourceID.Tag, Valid: report.SourceID.Tag != ""}
	result, err := rdb.db.Exec(sql1, report.SourceID.App, instance, tag, report.Created)
	if err != nil {
		return fmt.Errorf("failed to assign new group: %w", err)
	}
	groupID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to assign new group: %w", err)
	}
	report.GroupID = int(groupID)
	return nil
}

func (rdb *ReportsDatabase) InsertReport(report *general.Report) error {
	err := rdb.updateGroupID(report)
	if err == sql.ErrNoRows {
		if err := rdb.assignNewGroup(report); err != nil {
			return fmt.Errorf("failed to insert report: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to insert report: %w", err)
	}

	sql1 := "INSERT INTO conomi_report (report_group_id, severity, subject, body, args, created) VALUES (?,?,?,?,?,?)"
	entry, err := NewReportSQL(report)
	if err != nil {
		return fmt.Errorf("failed to insert report: %w", err)
	}
	log.Debug().Str("sql", sql1).Msg("going to INSERT report")
	result, err := rdb.db.Exec(sql1, entry.GroupID, entry.Severity, entry.Subject, entry.Body, entry.Args, entry.Created)
	if err != nil {
		return fmt.Errorf("failed to insert report: %w", err)
	}
	reportID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to insert report: %w", err)
	}
	report.ID = int(reportID)
	return nil
}

func (rdb *ReportsDatabase) ListReports(sourceID general.SourceID, resolved bool) ([]*general.Report, error) {
	whereParts := make([]string, 0, 4)
	whereValues := make([]any, 0, 3)
	if !resolved {
		whereParts = append(whereParts, "resolved_by_user_id IS NULL")
	}
	if sourceID.App != "" {
		whereParts = append(whereParts, "app = ?")
		whereValues = append(whereValues, sourceID.App)
	}
	if sourceID.Instance != "" {
		whereParts = append(whereParts, "instance = ?")
		whereValues = append(whereValues, sourceID.Instance)
	}
	if sourceID.Tag != "" {
		whereParts = append(whereParts, "tag = ?")
		whereValues = append(whereValues, sourceID.Tag)
	}
	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ") + " "
	}
	sql1 := "SELECT cr.id, crg.id, crg.app, crg.instance, crg.tag, cr.severity, cr.subject, cr.body, cr.args, cr.created, crg.resolved_by_user_id, us.user, crg.escalated " +
		"FROM conomi_report_group AS crg " +
		"JOIN conomi_report AS cr ON crg.id = cr.report_group_id " +
		"LEFT JOIN user AS us ON resolved_by_user_id = us.id " +
		whereClause +
		"ORDER BY created DESC"
	log.Debug().Str("sql", sql1).Msg("going to SELECT conomi_report")
	rows, err := rdb.db.Query(sql1, whereValues...)
	if err != nil {
		return nil, err
	}
	ans := make([]*general.Report, 0, 100)
	for rows.Next() {
		entry := &reportSQL{}
		err := rows.Scan(&entry.ID, &entry.GroupID, &entry.App, &entry.Instance, &entry.Tag, &entry.Severity, &entry.Subject, &entry.Body, &entry.Args, &entry.Created, &entry.ResolvedByUserID, &entry.ResolvedByUserName, &entry.Escalated)
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
	sql1 := "SELECT cr.id, crg.id, crg.app, crg.instance, crg.tag, cr.severity, cr.subject, cr.body, cr.args, cr.created, crg.resolved_by_user_id, us.user, crg.escalated " +
		"FROM conomi_report_group AS crg " +
		"JOIN conomi_report AS cr ON crg.id = cr.report_group_id " +
		"LEFT JOIN user AS us ON resolved_by_user_id = us.id " +
		"WHERE cr.id = ? LIMIT 1"
	log.Debug().Str("sql", sql1).Msgf("going to SELECT conomi_reports WHERE id = %d", reportID)
	entry := &reportSQL{}
	row := rdb.db.QueryRow(sql1, reportID)
	if err := row.Scan(&entry.ID, &entry.GroupID, &entry.App, &entry.Instance, &entry.Tag, &entry.Severity, &entry.Subject, &entry.Body, &entry.Args, &entry.Created, &entry.ResolvedByUserID, &entry.ResolvedByUserName, &entry.Escalated); err != nil {
		return nil, err
	}
	if err := entry.Severity.Validate(); err != nil {
		return nil, err
	}
	return entry.Export()
}

func (rdb *ReportsDatabase) EscalateGroup(groupID int) error {
	sql1 := "UPDATE conomi_report_group AS crg " +
		"SET crg.escalated = 1 " +
		"WHERE crg.resolved_by_user_id IS NULL AND crg.id = ?"
	log.Debug().Str("sql", sql1).Msgf("going to escalate group WHERE id = %d", groupID)
	_, err := rdb.db.Exec(sql1, groupID)
	if err != nil {
		return fmt.Errorf("failed to escalate group: %w", err)
	}
	return nil
}

func (rdb *ReportsDatabase) ResolveGroup(groupID int, userID int) error {
	sql1 := "UPDATE conomi_report_group AS crg " +
		"SET crg.resolved_by_user_id = ? " +
		"WHERE crg.resolved_by_user_id IS NULL AND crg.id = ?"
	log.Debug().Str("sql", sql1).Msgf("going to resolve group WHERE id = %d", groupID)
	_, err := rdb.db.Exec(sql1, userID, groupID)
	if err != nil {
		return fmt.Errorf("failed to resolve group: %w", err)
	}
	return nil
}

func (rdb *ReportsDatabase) GetOverview() ([]*general.ReportOverview, error) {
	sql1 := "SELECT crg.app, crg.instance, crg.tag, crg.escalated, " +
		"SUM(CASE WHEN cr.severity = ? THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN cr.severity = ? THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN cr.severity = ? THEN 1 ELSE 0 END), " +
		"SUM(CASE WHEN (cr.created > NOW() - INTERVAL 1 DAY) THEN 1 ELSE 0 END) AS recent, " +
		"crg.created, MAX(cr.created) " +
		"FROM conomi_report_group AS crg " +
		"JOIN conomi_report AS cr ON crg.id = cr.report_group_id " +
		"WHERE crg.resolved_by_user_id IS NULL " +
		"GROUP BY crg.app, crg.instance, crg.tag ORDER BY recent DESC, crg.app, crg.instance, crg.tag"
	log.Debug().Str("sql", sql1).Msg("going to count conomi_report WHERE resolved_by_user_id IS NULL")
	rows, err := rdb.db.Query(sql1, general.SeverityLevelCritical, general.SeverityLevelWarning, general.SeverityLevelInfo)
	if err != nil {
		return nil, err
	}
	ans := make([]*general.ReportOverview, 0, 100)
	for rows.Next() {
		count := &general.ReportOverview{}
		var instance, tag sql.NullString
		err := rows.Scan(&count.SourceID.App, &instance, &tag, &count.Escalated, &count.Critical, &count.Warning, &count.Info, &count.Recent, &count.Created, &count.Last)
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

func (rdb *ReportsDatabase) GetUserID(userName string) (int, error) {
	sql1 := "SELECT id FROM user WHERE user = ? LIMIT 1"
	log.Debug().Str("sql", sql1).Msg("going to get user id from name")
	row := rdb.db.QueryRow(sql1, userName)
	var userID int
	err := row.Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("failed to find user with username `%s`", userName)
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func NewReportsDatabase(db *sql.DB) *ReportsDatabase {
	return &ReportsDatabase{
		db:  db,
		ctx: context.Background(),
	}
}
