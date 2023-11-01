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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/czcorpus/conomi/general"
	"github.com/czcorpus/conomi/notifiers/common"
	"github.com/rs/zerolog/log"
)

type ZulipNotifierArgs struct {
	Server     string   `json:"server"`
	Sender     string   `json:"sender"`
	Token      string   `json:"token"`
	Type       string   `json:"type"`
	Recipients []string `json:"recipients"`
}

type zulipNotifier struct {
	version general.VersionInfo
	args    *ZulipNotifierArgs
	filter  common.FilterConf
	loc     *time.Location
}

func (zn *zulipNotifier) ShouldBeSent(message general.Report) bool {
	return zn.filter.IsFiltered(message)
}

func (zn *zulipNotifier) SendNotification(message general.Report) error {
	params := url.Values{}
	params.Set("type", zn.args.Type)
	for _, recipient := range zn.args.Recipients {
		params.Add("to", recipient)
	}
	completeMessage := strings.Join(
		[]string{
			fmt.Sprintf("***%s*: %s** (%s/%s)", strings.ToUpper(message.Level), message.Subject, message.App, message.Instance),
			"",
			message.Body,
		},
		"\n",
	)
	params.Add("content", completeMessage)

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
	req.Header.Set("User-Agent", fmt.Sprintf("CNKNotifier/%s-%s", zn.version.Version, zn.version.GitCommit))
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
	version general.VersionInfo,
	args *ZulipNotifierArgs,
	filter common.FilterConf,
	loc *time.Location,
) (common.Notifier, error) {
	log.Info().Msgf("creating zulip notifier of type `%s` with recipient(s) %s", args.Type, strings.Join(args.Recipients, ", "))
	notifier := &zulipNotifier{
		version: version,
		args:    args,
		filter:  filter,
		loc:     loc,
	}
	return notifier, nil
}
