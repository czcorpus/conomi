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

package common

import (
	"github.com/czcorpus/conomi/general"
)

type NotifierConf struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Args   map[string]any `json:"args"`
	Filter FilterConf     `json:"filter"`
}

type FilterConf struct {
	Levels []general.SeverityLevel `json:"levels"`
	Apps   []string                `json:"apps"`
}

func (fc *FilterConf) Validate() error {
	for _, level := range fc.Levels {
		if err := level.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func contains[T comparable](slice []T, item T) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func (f *FilterConf) IsFiltered(message *general.Report) bool {
	validLevel, validApp := true, true
	if f.Levels != nil {
		validLevel = contains(f.Levels, message.Severity)
	}
	if f.Apps != nil {
		validApp = contains(f.Apps, message.App)
	}
	return validLevel && validApp
}
