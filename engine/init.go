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

package engine

import (
	"database/sql"
	"fmt"
)

func initDatabase(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE conomi_report_group (
		id int(11) NOT NULL AUTO_INCREMENT,
		app varchar(50) NOT NULL,
		instance varchar(50),
		tag varchar(100),
		created datetime DEFAULT NOW() NOT NULL,
		escalated TINYINT(1) DEFAULT 0,
		resolved_by_user_id int DEFAULT NULL,
		PRIMARY KEY (id)
	)`)

	if err != nil {
		return fmt.Errorf("failed to CREATE table conomi_report_group: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE conomi_report (
		id int(11) NOT NULL AUTO_INCREMENT,
		report_group_id int(11) NOT NULL REFERENCES conomi_report_group(id),
		severity varchar(50) NOT NULL,
		subject text NOT NULL,
		body text NOT NULL,
		args json,
		created datetime DEFAULT NOW() NOT NULL,
		PRIMARY KEY (id)
	)`)

	if err != nil {
		return fmt.Errorf("failed to CREATE table conomi_reports: %w", err)
	}

	return nil
}
