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
	"text/template"
	"time"

	"github.com/czcorpus/cnc-gokit/mail"
	"github.com/czcorpus/conomi/general"
	"github.com/czcorpus/conomi/notifiers/common"
	"github.com/czcorpus/conomi/templates"
	"github.com/rs/zerolog/log"
)

type emailNotifier struct {
	info   general.GeneralInfo
	args   *mail.NotificationConf
	filter common.FilterConf
	loc    *time.Location
	tmpl   *template.Template
}

func (en *emailNotifier) ShouldBeSent(report general.Report) bool {
	return en.filter.IsFiltered(report)
}

func (en *emailNotifier) SendNotification(report general.Report) error {
	var message strings.Builder
	if err := en.tmpl.Execute(&message, templates.TemplateData{Report: report, Info: en.info}); err != nil {
		return err
	}
	return mail.SendNotification(
		en.args,
		en.loc,
		mail.FormattedNotification{
			Subject: fmt.Sprintf("%s: %s (%s/%s)", strings.ToUpper(report.Level), report.Subject, report.App, report.Instance),
			Divs:    []string{message.String()},
		},
	)
}

func NewEmailNotifier(
	info general.GeneralInfo,
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
	tmpl, err := templates.GetTemplate("email.gtpl")
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("creating e-mail notifier with recipient(s) %s", strings.Join(args.Recipients, ", "))
	notifier := &emailNotifier{
		info:   info,
		args:   args,
		filter: filter,
		loc:    loc,
		tmpl:   tmpl,
	}
	return notifier, nil
}
