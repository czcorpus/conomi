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

package reporting

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/czcorpus/conomi/engine"
	"github.com/czcorpus/conomi/general"
	"github.com/czcorpus/conomi/notifiers/common"
	"github.com/gin-gonic/gin"
)

type Actions struct {
	db *sql.DB
	n  []common.Notifier
}

func (a *Actions) PostReport(ctx *gin.Context) {
	var report general.Report
	if err := ctx.ShouldBindJSON(&report); err != nil {
		uniresp.WriteJSONErrorResponse(
			ctx.Writer, uniresp.NewActionErrorFrom(err), http.StatusBadRequest)
		return
	}
	rdb := engine.NewReportsDatabase(a.db)
	if err := rdb.InsertReport(report); err != nil {
		uniresp.WriteJSONErrorResponse(
			ctx.Writer, uniresp.NewActionErrorFrom(err), http.StatusInternalServerError)
		return
	}
	for _, notifier := range a.n {
		if notifier.ShouldBeSent(report) {
			if err := notifier.SendNotification(report); err != nil {
				uniresp.WriteJSONErrorResponse(
					ctx.Writer, uniresp.NewActionErrorFrom(err), http.StatusInternalServerError)
				return
			}
		}
	}
	uniresp.WriteJSONResponse(ctx.Writer, map[string]bool{"ok": true})
}

func (a *Actions) GetReports(ctx *gin.Context) {
	rdb := engine.NewReportsDatabase(a.db)
	reports, err := rdb.SelectReports()
	if err != nil {
		uniresp.WriteJSONErrorResponse(
			ctx.Writer, uniresp.NewActionErrorFrom(err), http.StatusInternalServerError)
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, reports)
}

func (a *Actions) ResolveReport(ctx *gin.Context) {
	reportIDString := ctx.Param("reportId")
	reportID, err := strconv.Atoi(reportIDString)
	if err != nil {
		uniresp.WriteJSONErrorResponse(
			ctx.Writer, uniresp.NewActionErrorFrom(err), http.StatusInternalServerError)
		return
	}
	rdb := engine.NewReportsDatabase(a.db)
	if err := rdb.ResolveReport(reportID); err != nil {
		uniresp.WriteJSONErrorResponse(
			ctx.Writer, uniresp.NewActionErrorFrom(err), http.StatusInternalServerError)
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, map[string]bool{"ok": true})
}

func (a *Actions) GetReport(ctx *gin.Context) {
	reportIDString := ctx.Param("reportId")
	reportID, err := strconv.Atoi(reportIDString)
	if err != nil {
		uniresp.WriteJSONErrorResponse(
			ctx.Writer, uniresp.NewActionErrorFrom(err), http.StatusInternalServerError)
		return
	}
	rdb := engine.NewReportsDatabase(a.db)
	reports, err := rdb.SelectReport(reportID)
	if err != nil {
		uniresp.WriteJSONErrorResponse(
			ctx.Writer, uniresp.NewActionErrorFrom(err), http.StatusInternalServerError)
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, reports)
}

func NewActions(db *sql.DB, n []common.Notifier) *Actions {
	return &Actions{
		db: db,
		n:  n,
	}
}
