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

package general

import (
	"fmt"
)

type SeverityLevel string

const (
	SeverityLevelInfo     SeverityLevel = "info"
	SeverityLevelWarning  SeverityLevel = "warning"
	SeverityLevelCritical SeverityLevel = "critical"
	SeverityLevelRecovery SeverityLevel = "recovery"
)

func (sl SeverityLevel) Validate() error {
	if !(sl == SeverityLevelInfo ||
		sl == SeverityLevelWarning ||
		sl == SeverityLevelCritical ||
		sl == SeverityLevelRecovery) {
		return fmt.Errorf(
			"invalid level `%s`, use: `%s`, `%s`, `%s` or `%s`",
			sl,
			SeverityLevelInfo,
			SeverityLevelWarning,
			SeverityLevelCritical,
			SeverityLevelRecovery,
		)
	}
	return nil
}

func (sl SeverityLevel) String() string {
	return string(sl)
}
