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
	"database/sql"
	"encoding/json"
	"time"

	"github.com/czcorpus/conomi/general"
)

// structure for simpler record data transformation between sql and go
type ReportSQL struct {
	ID               int
	App              string
	Instance         sql.NullString
	Tag              sql.NullString
	Severity         general.SeverityLevel
	Subject          string
	Body             string
	Args             sql.NullString
	Created          time.Time
	ResolvedByUserID sql.NullInt32
}

func (r *ReportSQL) Export() (*general.Report, error) {
	var args map[string]any = nil
	if r.Args.Valid {
		err := json.Unmarshal([]byte(r.Args.String), &args)
		if err != nil {
			return nil, err
		}
	}
	resolvedByUserID := -1
	if r.ResolvedByUserID.Valid {
		resolvedByUserID = int(r.ResolvedByUserID.Int32)
	}
	return &general.Report{
		SourceID: general.SourceID{
			App:      r.App,
			Instance: r.Instance.String,
			Tag:      r.Tag.String,
		},
		ID:               r.ID,
		Severity:         r.Severity,
		Subject:          r.Subject,
		Body:             r.Body,
		Args:             args,
		Created:          r.Created,
		ResolvedByUserID: resolvedByUserID,
	}, nil
}

func NewReportSQL(r general.Report) (*ReportSQL, error) {
	args := sql.NullString{Valid: false, String: ""}
	if r.Args != nil {
		value, err := json.Marshal(r.Args)
		if err != nil {
			return nil, err
		}
		args.Valid = true
		args.String = string(value)
	}
	return &ReportSQL{
		ID:               r.ID,
		App:              r.SourceID.App,
		Instance:         sql.NullString{Valid: r.SourceID.Instance != "", String: r.SourceID.Instance},
		Tag:              sql.NullString{Valid: r.SourceID.Tag != "", String: r.SourceID.Tag},
		Severity:         r.Severity,
		Subject:          r.Subject,
		Body:             r.Body,
		Args:             args,
		Created:          r.Created,
		ResolvedByUserID: sql.NullInt32{Valid: r.ResolvedByUserID != -1, Int32: int32(r.ResolvedByUserID)},
	}, nil
}
