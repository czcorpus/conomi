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

package cnf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/czcorpus/conomi/auth"
	"github.com/czcorpus/conomi/engine"
	"github.com/czcorpus/conomi/notifiers/common"
	"github.com/rs/zerolog/log"
)

const (
	dfltServerWriteTimeoutSecs = 30
	dfltLanguage               = "en"
	dfltTimeZone               = "Europe/Prague"
)

// Conf is a global configuration of the app
type Conf struct {
	ListenAddress          string                `json:"listenAddress"`
	ListenPort             int                   `json:"listenPort"`
	ClientDistDirPath      string                `json:"clientDistDirPath"`
	ClientAssetsDirPath    string                `json:"clientAssetsDirPath"`
	ServerReadTimeoutSecs  int                   `json:"serverReadTimeoutSecs"`
	ServerWriteTimeoutSecs int                   `json:"serverWriteTimeoutSecs"`
	LogFile                string                `json:"logFile"`
	LogLevel               logging.LogLevel      `json:"logLevel"`
	Language               string                `json:"language"`
	TimeZone               string                `json:"timeZone"`
	DB                     *engine.DBConf        `json:"db"`
	Notifiers              []common.NotifierConf `json:"notifiers"`
	PublicPath             string                `json:"publicPath"`
	Auth                   *auth.AuthConf        `json:"auth"`

	srcPath string
}

func (conf *Conf) IsDebugMode() bool {
	return conf.LogLevel == "debug"
}

func (conf *Conf) TimezoneLocation() *time.Location {
	// we can ignore the error here as we always call c.Validate()
	// first (which also tries to load the location and report possible
	// error)
	loc, _ := time.LoadLocation(conf.TimeZone)
	return loc
}

// GetSourcePath returns an absolute path of a file
// the config was loaded from.
func (conf *Conf) GetSourcePath() string {
	if filepath.IsAbs(conf.srcPath) {
		return conf.srcPath
	}
	var cwd string
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "[failed to get working dir]"
	}
	return filepath.Join(cwd, conf.srcPath)
}

func LoadConfig(path string) *Conf {
	if path == "" {
		log.Fatal().Msg("Cannot load config - path not specified")
	}
	rawData, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load config")
	}
	var conf Conf
	conf.srcPath = path
	err = json.Unmarshal(rawData, &conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load config")
	}
	return &conf
}

func ValidateAndDefaults(conf *Conf) {
	if conf.ServerWriteTimeoutSecs == 0 {
		conf.ServerWriteTimeoutSecs = dfltServerWriteTimeoutSecs
		log.Warn().Msgf(
			"serverWriteTimeoutSecs not specified, using default: %d",
			dfltServerWriteTimeoutSecs,
		)
	}
	if conf.ClientDistDirPath == "" {
		conf.ClientDistDirPath = "./dist"
		log.Warn().
			Str("default", conf.ClientDistDirPath).
			Msg("clientDistDir not specified, using default")
	}
	if conf.ClientAssetsDirPath == "" {
		conf.ClientAssetsDirPath = "./assets"
		log.Warn().
			Str("default", conf.ClientAssetsDirPath).
			Msg("clientAssetsDirPath not specified, using default")
	}
	if conf.Language == "" {
		conf.Language = dfltLanguage
		log.Warn().Msgf("language not specified, using default: %s", conf.Language)
	}
	if conf.TimeZone == "" {
		log.Warn().
			Str("default", dfltTimeZone).
			Msg("time zone not specified, using default")
	}
	if _, err := time.LoadLocation(conf.TimeZone); err != nil {
		log.Fatal().Err(err).Msg("invalid time zone")
	}
	for _, notifier := range conf.Notifiers {
		if err := notifier.Filter.Validate(); err != nil {
			log.Fatal().Err(err).Msg("invalid filter")
		}
	}
}
