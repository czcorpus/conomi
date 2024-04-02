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
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/czcorpus/cnc-gokit/mail"
	"github.com/czcorpus/conomi/general"
	"github.com/czcorpus/conomi/notifiers/common"
	"github.com/czcorpus/conomi/reporting/content"
	"github.com/czcorpus/conomi/templates"
	"github.com/rs/zerolog/log"
)

type emailNotifier struct {
	name   string
	info   general.GeneralInfo
	args   *mail.NotificationConf
	filter common.FilterConf
	loc    *time.Location
	tmpl   *template.Template
}

func (en *emailNotifier) ShouldBeSent(report *general.Report) bool {
	return en.filter.IsFiltered(report)
}

func (en *emailNotifier) SendNotification(report *general.Report) error {
	var message strings.Builder
	report.Body = content.MarkdownToHTML(report.Body)
	if err := en.tmpl.Execute(
		&message,
		templates.NotificationTemplateData{
			NotifierName: en.name,
			Report:       *report,
			Info:         en.info,
		},
	); err != nil {
		return fmt.Errorf("failed to evaluate notification template: %w", err)
	}
	subject := strings.ToUpper(report.Severity.String()) + ": " + report.Subject
	if report.Escalated {
		subject = "[ESCALATED] " + subject
	}
	if len(report.SourceID.Instance) > 0 {
		subject += " (" + report.SourceID.App + "/" + report.SourceID.Instance + ")"
	} else {
		subject += " (" + report.SourceID.App + ")"
	}
	return mail.SendNotification(
		en.args,
		en.loc,
		mail.FormattedNotification{
			Subject: subject,
			Divs:    []string{message.String()},
		},
	)
}

func NewEmailNotifier(
	conf *common.NotifierConf,
	loc *time.Location,
	info general.GeneralInfo,
	args *mail.NotificationConf,
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
	tmpl, err := templates.GetTemplate(filepath.Join(conf.TplDirPath, "email.gtpl"))
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("creating e-mail notifier `%s` with recipient(s) %v", conf.Name, args.Recipients)
	notifier := &emailNotifier{
		name:   conf.Name,
		info:   info,
		args:   args,
		filter: conf.Filter,
		loc:    loc,
		tmpl:   tmpl,
	}
	return notifier, nil
}
