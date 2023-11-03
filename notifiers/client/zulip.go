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

package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/czcorpus/conomi/general"
	"github.com/czcorpus/conomi/notifiers/common"
	"github.com/czcorpus/conomi/templates"
	"github.com/rs/zerolog/log"
)

type ZulipNotifierArgs struct {
	Server     string   `json:"server"`
	Sender     string   `json:"sender"`
	Token      string   `json:"token"`
	Type       string   `json:"type"`
	Recipients []string `json:"recipients"`
	Topic      string   `json:"topic"`
}

type zulipNotifier struct {
	name   string
	info   general.GeneralInfo
	args   *ZulipNotifierArgs
	filter common.FilterConf
	loc    *time.Location
	tmpl   *template.Template
}

func (zn *zulipNotifier) ShouldBeSent(report *general.Report) bool {
	return zn.filter.IsFiltered(report)
}

func (zn *zulipNotifier) SendNotification(report *general.Report) error {
	params := url.Values{}
	params.Set("type", zn.args.Type)
	if zn.args.Type == "stream" {
		params.Set("to", zn.args.Recipients[0])
		params.Set("topic", zn.args.Topic)
	} else {
		params.Set("to", strings.Join(zn.args.Recipients, ","))
	}

	var message strings.Builder
	if err := zn.tmpl.Execute(
		&message,
		templates.TemplateData{
			NotifierName: zn.name,
			Report:       *report,
			Info:         zn.info,
		},
	); err != nil {
		return err
	}
	params.Add("content", message.String())

	zURL, err := url.Parse(zn.args.Server)
	if err != nil {
		return err
	}
	zURL = zURL.JoinPath("api", "v1", "messages")
	zURL.RawQuery = params.Encode()

	req, err := http.NewRequest("POST", zURL.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("CNKNotifier/%s-%s", zn.info.Build.Version, zn.info.Build.GitCommit))
	req.SetBasicAuth(zn.args.Sender, zn.args.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Debug().Bytes("response", body).Msg("performed zulip post")
	return nil
}

func NewZulipNotifier(
	name string,
	info general.GeneralInfo,
	args *ZulipNotifierArgs,
	filter common.FilterConf,
	loc *time.Location,
) (common.Notifier, error) {
	switch args.Type {
	case "direct":
		if len(args.Recipients) == 0 {
			return nil, errors.New("zulip `direct` type requires at least one recipient")
		}
	case "stream":
		if len(args.Recipients) != 1 {
			return nil, errors.New("zulip `stream` type requires exactly one recipient")
		}
		if len(args.Topic) == 0 {
			return nil, errors.New("zulip `stream` type requires specified topic")
		}
	default:
		return nil, fmt.Errorf("unknown zulip type `%s`, use `direct` or `stream`", args.Type)
	}
	tmpl, err := templates.GetTemplate("zulip.gtpl")
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("creating zulip notifier `%s` of type `%s` with recipient(s) %v > %s", name, args.Type, args.Recipients, args.Topic)
	notifier := &zulipNotifier{
		name:   name,
		info:   info,
		args:   args,
		filter: filter,
		loc:    loc,
		tmpl:   tmpl,
	}
	return notifier, nil
}
