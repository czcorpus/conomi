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

package templates

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/czcorpus/conomi/general"
)

type NotificationTemplateData struct {
	NotifierName string
	Info         general.GeneralInfo
	Report       general.Report
}

func GetTemplate(absPath string) (*template.Template, error) {
	templateFunc := template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}
	return template.New(filepath.Base(absPath)).Funcs(templateFunc).ParseFiles(absPath)
}
