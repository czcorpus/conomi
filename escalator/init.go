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

package escalator

import (
	"database/sql"
	"fmt"

	"github.com/czcorpus/conomi/engine"
	"github.com/czcorpus/conomi/general"
)

const escalateWarningCount = 10

type Escalator struct {
	counts map[string]*general.ReportCount
	db     *sql.DB
}

func (e *Escalator) makeKey(app, instance, tag string) string {
	return fmt.Sprintf("%s:%s:%s", app, instance, tag)
}

func (e *Escalator) Set(count *general.ReportCount) {
	key := e.makeKey(count.App, count.Instance, count.Tag)
	e.counts[key] = count
}

func (e *Escalator) Add(report *general.Report) bool {
	key := e.makeKey(report.App, report.Instance, report.Tag)
	count, ok := e.counts[key]
	if !ok {
		count = &general.ReportCount{App: report.App, Instance: report.Instance, Tag: report.Tag}
		e.counts[key] = count
	}
	switch report.Severity {
	case general.SeverityLevelCritical:
		count.Critical += 1
	case general.SeverityLevelWarning:
		count.Warning += 1
	case general.SeverityLevelInfo:
		count.Info += 1
	}

	lastEscalated := count.Escalated
	count.Escalated = count.Critical > 0 || count.Warning > escalateWarningCount
	// return if report just escalated
	return !lastEscalated && count.Escalated
}

func (e *Escalator) IsEscalated(report *general.Report) bool {
	key := e.makeKey(report.App, report.Instance, report.Tag)
	count, ok := e.counts[key]
	if ok {
		return count.Escalated
	}
	return false
}

func (e *Escalator) Reload() error {
	rdb := engine.NewReportsDatabase(e.db)
	counts, err := rdb.GetReportCounts()
	if err != nil {
		return err
	}
	e.counts = make(map[string]*general.ReportCount)
	for _, count := range counts {
		e.Set(count)
	}
	return nil
}

func NewEscalator(sqlDB *sql.DB) (*Escalator, error) {
	escalator := Escalator{db: sqlDB}
	if err := escalator.Reload(); err != nil {
		return nil, err
	}
	return &escalator, nil
}
