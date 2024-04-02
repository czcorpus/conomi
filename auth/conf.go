// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
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

package auth

type AuthConf struct {
	RemoteUserHeader string      `json:"remoteUserHeader"`
	ToolbarURL       string      `json:"toolbarUrl"`
	CookieSID        string      `json:"cookieSid"`
	CookieAt         string      `json:"cookieAt"`
	CookieRmme       string      `json:"cookieRmme"`
	CookieLang       string      `json:"cookieLang"`
	APITokens        []TokenConf `json:"apiTokens"`

	// UnsafeForceFallbackUser can be used for development and testing
	// purposes along with APITokens and RemoteUserHeader to allow
	// clients authentication into actions without actually having
	// a proper token. This should not be used in production!!!
	UnsafeForceFallbackUser string `json:"UNSAFE_setFallbackUser"`
}

type TokenConf struct {
	APITokenHash string `json:"apiTokenHash"`
	ClientID     string `json:"clientId"`
}
