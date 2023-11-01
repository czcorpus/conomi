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
	goMail "net/mail"
	"strings"
	"time"

	"github.com/czcorpus/cnc-gokit/mail"
	"github.com/czcorpus/conomi/general"
	"github.com/czcorpus/conomi/notifiers/common"
	"github.com/rs/zerolog/log"
)

type emailNotifier struct {
	version general.VersionInfo
	args    *mail.NotificationConf
	filter  common.FilterConf
	loc     *time.Location
}

func (en *emailNotifier) ShouldBeSent(message general.Report) bool {
	return en.filter.IsFiltered(message)
}

func (en *emailNotifier) SendNotification(message general.Report) error {
	return mail.SendNotification(en.args, en.loc, mail.Notification{
		Subject:    fmt.Sprintf("%s: %s (%s/%s)", strings.ToUpper(message.Level), message.Subject, message.App, message.Instance),
		Paragraphs: []string{message.Body},
	})
}

func NewEmailNotifier(
	version general.VersionInfo,
	args *mail.NotificationConf,
	filter common.FilterConf,
	loc *time.Location,
) (common.Notifier, error) {
	if args.Sender == "" {
		return nil, fmt.Errorf("e-mail sender not set")
	}
	validated := append([]string{args.Sender}, args.Recipients...)
	for _, addr := range validated {
		if _, err := goMail.ParseAddress(addr); err != nil {
			return nil, fmt.Errorf("incorrect e-mail address %s: %s", addr, err)
		}
	}
	log.Info().Msgf("creating e-mail notifier with recipient(s) %s", strings.Join(args.Recipients, ", "))
	notifier := &emailNotifier{
		version: version,
		args:    args,
		filter:  filter,
		loc:     loc,
	}
	return notifier, nil
}
