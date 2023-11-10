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

package notifiers

import (
	"fmt"
	"time"

	"github.com/czcorpus/cnc-gokit/mail"
	"github.com/czcorpus/conomi/general"
	"github.com/czcorpus/conomi/notifiers/client"
	"github.com/czcorpus/conomi/notifiers/common"
	"github.com/mitchellh/mapstructure"
)

func clientsFactory(
	info general.GeneralInfo,
	notifiersConf []common.NotifierConf,
	loc *time.Location,
) ([]common.Notifier, error) {
	clients := make([]common.Notifier, len(notifiersConf))
	for i, conf := range notifiersConf {
		switch conf.Type {
		case "email":
			var emailConf mail.NotificationConf
			err := mapstructure.Decode(conf.Args, &emailConf)
			if err != nil {
				return nil, fmt.Errorf("invalid email notifier conf: %s", err)
			}
			clients[i], err = client.NewEmailNotifier(&conf, loc, info, &emailConf)
			if err != nil {
				return nil, err
			}
		case "zulip":
			var zulipConf client.ZulipNotifierArgs
			err := mapstructure.Decode(conf.Args, &zulipConf)
			if err != nil {
				return nil, fmt.Errorf("invalid zulip notifier conf: %s", err)
			}
			clients[i], err = client.NewZulipNotifier(&conf, loc, info, &zulipConf)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown notifier type %s", conf.Type)
		}
	}
	return clients, nil
}

type Notifiers struct {
	notifiers []common.Notifier
}

func (n *Notifiers) SendNotifications(report *general.Report) error {
	for _, client := range n.notifiers {
		if client.ShouldBeSent(report) {
			if err := client.SendNotification(report); err != nil {
				return err
			}
		}
	}
	return nil
}

func NewNotifiers(info general.GeneralInfo, notifiersConf []common.NotifierConf, loc *time.Location) (*Notifiers, error) {
	clients, err := clientsFactory(info, notifiersConf, loc)
	if err != nil {
		return nil, err
	}
	return &Notifiers{notifiers: clients}, nil
}
