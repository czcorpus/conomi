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

package general

import (
	"time"
)

type SourceID struct {
	App      string `json:"app"`
	Instance string `json:"instance"`
	Tag      string `json:"tag"`
}

type Report struct {
	SourceID           SourceID       `json:"sourceId"`
	ID                 int            `json:"id"`
	GroupID            int            `json:"groupId"`
	Severity           SeverityLevel  `json:"severity"`
	Subject            string         `json:"subject"`
	Body               string         `json:"body"`
	Args               map[string]any `json:"args"`
	Created            time.Time      `json:"created"`
	ResolvedByUserID   int            `json:"resolvedByUserId"` // for empty user we use value -1
	ResolvedByUserName string         `json:"resolvedByUserName"`
	Escalated          bool           `json:"escalated"`
}

type ReportOverview struct {
	SourceID  SourceID  `json:"sourceId"`
	Escalated bool      `json:"escalated"`
	Critical  int       `json:"critical"`
	Warning   int       `json:"warning"`
	Info      int       `json:"info"`
	Recent    int       `json:"recent"`
	Created   time.Time `json:"created"`
	Last      time.Time `json:"last"`
}
