// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
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

package client

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/czcorpus/conomi/auth"
	"github.com/czcorpus/conomi/general"
	"github.com/rs/zerolog/log"
)

type ConomiClientConf struct {
	Server   string `json:"server"`
	App      string `json:"app"`
	Instance string `json:"instance"`
	APIToken string `json:"apiToken"`
}

type ConomiClient struct {
	conf ConomiClientConf
}

type conomiReport struct {
	SourceID general.SourceID      `json:"sourceId"`
	Severity general.SeverityLevel `json:"severity"`
	Subject  string                `json:"subject"`
	Body     string                `json:"body"`
	Args     map[string]any        `json:"args"`
}

func (cc *ConomiClient) Ping() error {
	reportURL, err := url.JoinPath(cc.conf.Server, "api", "ping")
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", reportURL, &bytes.Buffer{})
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	h := sha256.New()
	h.Write([]byte(cc.conf.APIToken))
	req.Header.Set(auth.ApiAuthTokenHTTPHeader, fmt.Sprintf("%x", h.Sum(nil)))

	resp, err := http.DefaultClient.Do(req)
	log.Debug().
		Err(err).
		Msg("sent Conomi ping via HTTP")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var msg map[string]bool
	err = json.Unmarshal(respBody, &msg)
	if err != nil {
		return err
	}
	log.Debug().Str("response", string(respBody)).Msg("conomi post performed")
	if msg["ok"] {
		return nil
	}
	return errors.New("obtained value != `ok`")
}

func (cc *ConomiClient) SendReport(severity general.SeverityLevel, subject string, body string, opts ...ReportOption) error {
	reportURL, err := url.JoinPath(cc.conf.Server, "api", "report")
	if err != nil {
		return err
	}
	report := conomiReport{
		SourceID: general.SourceID{
			App:      cc.conf.App,
			Instance: cc.conf.Instance,
		},
		Severity: severity,
		Subject:  subject,
		Body:     body,
	}
	for _, opt := range opts {
		opt(&report)
	}
	payload, err := json.Marshal(report)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", reportURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	h := sha256.New()
	h.Write([]byte(cc.conf.APIToken))
	req.Header.Set(auth.ApiAuthTokenHTTPHeader, fmt.Sprintf("%x", h.Sum(nil)))

	resp, err := http.DefaultClient.Do(req)
	log.Debug().
		Err(err).
		Str("severity", string(severity)).
		Str("subject", subject).
		Str("app", report.SourceID.App).
		Str("instance", report.SourceID.Instance).
		Str("tag", report.SourceID.Tag).
		Any("args", report.Args).
		Msg("sent Conomi report via HTTP")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Debug().Str("response", string(respBody)).Msg("conomi post performed")
	return nil
}

func NewConomiClient(conf ConomiClientConf) *ConomiClient {
	return &ConomiClient{
		conf: conf,
	}
}
