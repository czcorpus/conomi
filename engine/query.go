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
	"encoding/json"

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
	sql1 := "INSERT INTO conomi_reports (app, instance, level, subject, body, args, created) VALUES (?,?,?,?,?,?,?)"
	log.Debug().Str("sql", sql1).Msg("going to INSERT report")
	instance := sql.NullString{
		String: report.Instance,
		Valid:  len(report.Instance) > 0,
	}
	var args sql.NullString
	if report.Args != nil {
		argsJSON, err := json.Marshal(report.Args)
		if err != nil {
			return -1, err
		}
		args.String = string(argsJSON)
		args.Valid = true
	}
	result, err := rdb.db.Exec(sql1, report.App, instance, report.Level, report.Subject, report.Body, args, report.Created)
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
	sql1 := "SELECT id, app, instance, level, subject, body, args, created, resolved_by_user_id " +
		"FROM conomi_reports " +
		"WHERE resolved_by_user_id IS NULL"
	log.Debug().Str("sql", sql1).Msg("going to SELECT conomi_reports WHERE resolved_by_user_id IS NULL")
	rows, err := rdb.db.Query(sql1)
	if err != nil {
		return []*general.Report{}, err
	}
	ans := make([]*general.Report, 0, 100)
	for rows.Next() {
		var resolvedByUserID sql.NullInt32
		var instance, args sql.NullString
		item := &general.Report{ResolvedByUserID: -1}
		err := rows.Scan(&item.ID, &item.App, &instance, &item.Level, &item.Subject, &item.Body, &args, &item.Created, &resolvedByUserID)
		if err != nil {
			return ans, err
		}
		if resolvedByUserID.Valid {
			item.ResolvedByUserID = int(resolvedByUserID.Int32)
		}
		item.Instance = instance.String
		if args.Valid {
			err = json.Unmarshal([]byte(args.String), &item.Args)
			if err != nil {
				return ans, err
			}
		}
		ans = append(ans, item)
	}
	return ans, nil
}

func (rdb *ReportsDatabase) SelectReport(reportID int) (*general.Report, error) {
	sql1 := "SELECT id, app, instance, level, subject, body, args, created, resolved_by_user_id " +
		"FROM conomi_reports " +
		"WHERE id = ? LIMIT 1"
	log.Debug().Str("sql", sql1).Msgf("going to SELECT conomi_reports WHERE id = %d", reportID)
	var resolvedByUserID sql.NullInt32
	var instance, args sql.NullString
	item := &general.Report{ResolvedByUserID: -1}
	row := rdb.db.QueryRow(sql1, reportID)
	err := row.Scan(&item.ID, &item.App, &instance, &item.Level, &item.Subject, &item.Body, &args, &item.Created, &resolvedByUserID)
	if err != nil {
		return nil, err
	}
	if resolvedByUserID.Valid {
		item.ResolvedByUserID = int(resolvedByUserID.Int32)
	}
	item.Instance = instance.String
	if args.Valid {
		err = json.Unmarshal([]byte(args.String), &item.Args)
		if err != nil {
			return item, err
		}
	}
	return item, nil
}

func (rdb *ReportsDatabase) ResolveReport(reportID int, userID int) error {
	sql1 := "UPDATE conomi_reports SET resolved_by_user_id = ? WHERE id = ? AND resolved_by_user_id IS NULL"
	log.Debug().Str("sql", sql1).Msgf("going to resolve report WHERE id = %d", reportID)
	_, err := rdb.db.Exec(sql1, userID, reportID)
	return err
}

func NewReportsDatabase(db *sql.DB) *ReportsDatabase {
	return &ReportsDatabase{
		db:  db,
		ctx: context.Background(),
	}
}
