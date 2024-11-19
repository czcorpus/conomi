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

package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/czcorpus/conomi/auth"
	"github.com/czcorpus/conomi/engine"
	"github.com/czcorpus/conomi/escalator"
	"github.com/czcorpus/conomi/general"
	"github.com/czcorpus/conomi/notifiers"
	"github.com/czcorpus/conomi/reporting"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/czcorpus/conomi/cnf"
)

var (
	version   string
	buildDate string
	gitCommit string
)

func runApiServer(
	info general.GeneralInfo,
	conf *cnf.Conf,
	syscallChan chan os.Signal,
	exitEvent chan os.Signal,
	sqlDB *sql.DB,
) error {
	if !conf.Logging.Level.IsDebugMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(logging.GinMiddleware())
	engine.Use(auth.Authenticate(conf.Auth, conf.PublicPath))
	engine.NoMethod(uniresp.NoMethodHandler)
	engine.NoRoute(uniresp.NotFoundHandler)

	n, err := notifiers.NewNotifiers(info, conf.Notifiers, conf.TimezoneLocation())
	if err != nil {
		return fmt.Errorf("failed to instantiate notifiers: %w", err)
	}
	e, err := escalator.NewEscalator(sqlDB, n)
	if err != nil {
		return fmt.Errorf("failed to instantiate escalator: %w", err)
	}
	r := reporting.NewActions(conf.TimezoneLocation(), sqlDB, n, e)
	api := engine.Group("/api")
	api.Use(uniresp.AlwaysJSONContentType())
	api.Use(auth.AbortUnauthorized())
	api.GET("/ping", r.Ping)
	api.POST("/report", r.PostReport)
	api.GET("/report/:reportId", r.GetReport)
	api.POST("/resolve/:groupId", r.ResolveGroup)
	api.GET("/reports", r.GetReports)
	api.GET("/sources", r.GetSources)
	api.GET("/overview", r.GetOverview)

	engine.LoadHTMLFiles(filepath.Join(conf.ClientDistDirPath, "index.html"))
	ui := engine.Group("/ui")
	ui.StaticFS("/assets", http.Dir(conf.ClientAssetsDirPath))
	ui.Static("/js", filepath.Join(conf.ClientDistDirPath, "js"))
	ui.Static("/css", filepath.Join(conf.ClientDistDirPath, "css"))
	uiHandler := func(c *gin.Context) {
		params := gin.H{}
		toolbar, exists := c.Get("toolbar")
		if exists {
			tb, tbOk := toolbar.(map[string]interface{})
			if tbOk {
				toolbarHTML, htmlOk := tb["html"].(string)
				if htmlOk {
					params["toolbarHTML"] = template.HTML(toolbarHTML)
					params["toolbarStyles"] = tb["styles"]
					params["toolbarScripts"] = tb["scripts"]
				}
			}
		}
		auth, exists := c.Get("authenticated")
		if exists {
			if auth == true {
				userName, exists := c.Get("userName")
				if exists {
					params["userName"] = userName
				}
			} else {
				params["errorMsg"] = "Unauthorized"
				c.HTML(http.StatusUnauthorized, "index.html", params)
				return
			}

		}
		c.HTML(http.StatusOK, "index.html", params)
	}
	ui.GET("/", uiHandler)
	ui.GET("/list", uiHandler)
	ui.GET("/detail", uiHandler)

	log.Info().Msgf("starting to listen at %s:%d", conf.ListenAddress, conf.ListenPort)
	srv := &http.Server{
		Handler:      engine,
		Addr:         fmt.Sprintf("%s:%d", conf.ListenAddress, conf.ListenPort),
		WriteTimeout: time.Duration(conf.ServerWriteTimeoutSecs) * time.Second,
		ReadTimeout:  time.Duration(conf.ServerReadTimeoutSecs) * time.Second,
	}
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Error().Err(err).Msg("")
		}
		syscallChan <- syscall.SIGTERM
	}()

	<-exitEvent
	r.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = srv.Shutdown(ctx)
	if err != nil {
		log.Info().Err(err).Msg("Shutdown request error")
	}
	return nil
}

func main() {
	build := general.Build{
		Version:   version,
		BuildDate: buildDate,
		GitCommit: gitCommit,
	}
	createTables := flag.Bool("create-tables", false, "Force creating db tables (requires proper permissions)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Conomi - CNC Notification Middleware\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options] start [config.json]\n\t", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "%s hashtoken [token]\n\t", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "%s [options] version\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	action := flag.Arg(0)
	if action == "version" {
		fmt.Printf(
			"Conomi - CNC Notification Middleware\n%s\nbuild date: %s\nlast commit: %s\n",
			build.Version, build.BuildDate, build.GitCommit,
		)
		return

	} else if action == "hashtoken" {
		h := sha256.New()
		h.Write([]byte(flag.Arg(1)))
		fmt.Printf("%x\n", h.Sum(nil))
		return

	} else if action != "test" && action != "start" {
		fmt.Println("unknown action ", action)
		os.Exit(1)
		return
	}
	conf := cnf.LoadConfig(flag.Arg(1))
	info := general.GeneralInfo{
		Build:      build,
		PublicPath: conf.PublicPath,
	}

	if action == "test" {
		cnf.ValidateAndDefaults(conf)
		log.Info().Msg("config OK")
		return

	} else {
		logging.SetupLogging(conf.Logging)
	}
	log.Info().Msg("Starting Conomi")
	cnf.ValidateAndDefaults(conf)
	syscallChan := make(chan os.Signal, 1)
	signal.Notify(syscallChan, os.Interrupt)
	signal.Notify(syscallChan, syscall.SIGTERM)
	exitEvent := make(chan os.Signal)

	go func() {
		evt := <-syscallChan
		exitEvent <- evt
		close(exitEvent)
	}()

	switch action {
	case "start":
		sqlDB, err := engine.Open(conf.DB, *createTables)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open database connection")
		}
		err = runApiServer(info, conf, syscallChan, exitEvent, sqlDB)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to run api server")
		}
	default:
		log.Fatal().Msgf("Unknown action %s", action)
	}

}
