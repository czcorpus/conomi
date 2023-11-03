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
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/czcorpus/conomi/general"
	"github.com/rs/zerolog/log"
)

type ConomiClientConf struct {
	Server   string `json:"server"`
	App      string `json:"app"`
	Instance string `json:"instance"`
}

type ConomiClient struct {
	conf ConomiClientConf
}

type conomiReport struct {
	App      string                `json:"app"`
	Instance string                `json:"instance"`
	Tag      string                `json:"tag"`
	Severity general.SeverityLevel `json:"level"`
	Subject  string                `json:"subject"`
	Body     string                `json:"body"`
	Args     map[string]any        `json:"args"`
}

func (cc *ConomiClient) SendReport(severity general.SeverityLevel, subject string, body string, opts ...ReportOption) error {
	reportURL, err := url.JoinPath(cc.conf.Server, "report")
	if err != nil {
		return err
	}
	report := conomiReport{
		App:      cc.conf.App,
		Instance: cc.conf.Instance,
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

	resp, err := http.DefaultClient.Do(req)
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