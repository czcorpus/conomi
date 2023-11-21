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
	"time"

	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/czcorpus/conomi/auth"
	"github.com/czcorpus/conomi/engine"
	"github.com/czcorpus/conomi/escalator"
	"github.com/czcorpus/conomi/general"
	"github.com/czcorpus/conomi/notifiers"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Actions struct {
	loc *time.Location
	db  *sql.DB
	n   *notifiers.Notifiers
	e   *escalator.Escalator
}

func (a *Actions) autoResolve(ctx *gin.Context, rdb *engine.ReportsDatabase, groupID int) error {
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return err
	}
	if err := rdb.ResolveGroup(groupID, userID); err != nil {
		return err
	}
	if err := a.e.Reload(); err != nil {
		return err
	}
	return nil
}

func (a *Actions) PostReport(ctx *gin.Context) {
	report := general.Report{ResolvedByUserID: -1, Created: time.Now().In(a.loc)}
	if err := ctx.ShouldBindJSON(&report); err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusBadRequest)
		return
	}
	log.Debug().
		Str("severity", string(report.Severity)).
		Str("subject", report.Subject).
		Str("app", report.SourceID.App).
		Str("instance", report.SourceID.Instance).
		Str("tag", report.SourceID.Tag).
		Any("args", report.Args).
		Msg("Obtained report via HTTP API")
	if err := report.Severity.Validate(); err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusBadRequest)
		return
	}
	rdb := engine.NewReportsDatabase(a.db)
	reportID, groupID, err := rdb.InsertReport(report)
	if err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusInternalServerError)
		return
	}
	report.ID = reportID
	report.GroupID = groupID
	if report.Severity == general.SeverityLevelRecovery {
		if err := a.autoResolve(ctx, rdb, groupID); err != nil {
			log.Error().AnErr("error", err).Msg("auto resolve failed")
		}
	}
	if err := a.e.HandleReport(&report); err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusInternalServerError)
		return
	}
	if err := a.n.SendNotifications(&report); err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusInternalServerError)
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, report)
}

func (a *Actions) GetReports(ctx *gin.Context) {
	sourceID := general.SourceID{
		App:      ctx.Request.URL.Query().Get("app"),
		Instance: ctx.Request.URL.Query().Get("instance"),
		Tag:      ctx.Request.URL.Query().Get("tag"),
	}
	rdb := engine.NewReportsDatabase(a.db)
	reports, err := rdb.ListReports(sourceID)
	if err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusInternalServerError)
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, reports)
}

func (a *Actions) ResolveGroup(ctx *gin.Context) {
	groupIDString := ctx.Param("groupId")
	groupID, err := strconv.Atoi(groupIDString)
	if err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusBadRequest)
		return
	}
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusInternalServerError)
		return
	}
	rdb := engine.NewReportsDatabase(a.db)
	err = rdb.ResolveGroup(groupID, userID)
	if err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusInternalServerError)
		return
	}
	if err := a.e.Reload(); err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusInternalServerError)
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, map[string]any{"ok": true})
}

func (a *Actions) GetReport(ctx *gin.Context) {
	reportIDString := ctx.Param("reportId")
	reportID, err := strconv.Atoi(reportIDString)
	if err != nil {
		uniresp.RespondWithErrorJSON(
			ctx, err, http.StatusInternalServerError)
		return
	}
	rdb := engine.NewReportsDatabase(a.db)
	report, err := rdb.SelectReport(reportID)
	if err != nil {
		if err == sql.ErrNoRows {
			uniresp.RespondWithErrorJSON(
				ctx, err, http.StatusNotFound)
		} else {
			uniresp.RespondWithErrorJSON(
				ctx, err, http.StatusInternalServerError)
		}
		return
	}
	if ctx.Query("md-to-html") == "1" {
		report.Body = mdToHTML(report.Body)
	}
	uniresp.WriteJSONResponse(ctx.Writer, report)
}

func (a *Actions) GetSources(ctx *gin.Context) {
	rdb := engine.NewReportsDatabase(a.db)
	filters, err := rdb.GetSources()
	if err != nil {
		if err == sql.ErrNoRows {
			uniresp.RespondWithErrorJSON(
				ctx, err, http.StatusNotFound)
		} else {
			uniresp.RespondWithErrorJSON(
				ctx, err, http.StatusInternalServerError)
		}
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, filters)
}

func (a *Actions) GetReportCounts(ctx *gin.Context) {
	rdb := engine.NewReportsDatabase(a.db)
	counts, err := rdb.GetReportCounts()
	if err != nil {
		if err == sql.ErrNoRows {
			uniresp.RespondWithErrorJSON(
				ctx, err, http.StatusNotFound)
		} else {
			uniresp.RespondWithErrorJSON(
				ctx, err, http.StatusInternalServerError)
		}
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, counts)
}

func NewActions(loc *time.Location, db *sql.DB, n *notifiers.Notifiers, e *escalator.Escalator) *Actions {
	return &Actions{
		loc: loc,
		db:  db,
		n:   n,
		e:   e,
	}
}
