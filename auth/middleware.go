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

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	ApiAuthTokenHTTPHeader = "x-conomi-token"
)

func Authenticate(conf *AuthConf, publicPath string) gin.HandlerFunc {
	if conf == nil {
		return func(ctx *gin.Context) {
			ctx.Set("authenticated", true)
			ctx.Next()
		}
	} else if conf.RemoteUserHeader != "" || conf.APITokenHash != "" {
		return func(ctx *gin.Context) {
			ctx.Set("authenticated", false)
			remoteUser := ctx.Request.Header.Get(conf.RemoteUserHeader)
			if remoteUser != "" {
				ctx.Set("authenticated", true)
				ctx.Set("userName", remoteUser)

			} else {
				log.Debug().
					Str("expected", conf.APITokenHash).
					Str("obtained", ctx.Request.Header.Get(ApiAuthTokenHTTPHeader)).
					Msg("testing token authentication")
				ctx.Set("authenticated", ctx.Request.Header.Get(ApiAuthTokenHTTPHeader) == conf.APITokenHash)
			}
			ctx.Next()
		}
	}
	return func(ctx *gin.Context) {
		ctx.Set("authenticated", false)
		cookieSID, err := ctx.Cookie(conf.CookieSID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		cookieAt, err := ctx.Cookie(conf.CookieAt)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		cookieRmme, err := ctx.Cookie(conf.CookieRmme)
		if err != nil {
			cookieRmme = "0"
		}
		cookieLang, err := ctx.Cookie(conf.CookieLang)
		if err != nil {
			cookieLang = "en"
		}

		params := url.Values{}
		params.Add("sid", cookieSID)
		params.Add("at", cookieAt)
		params.Add("rmme", cookieRmme)
		params.Add("lang", cookieLang)
		params.Add("current", "conomi")
		params.Add("continue", publicPath+ctx.FullPath())
		query := params.Encode()

		req, err := http.NewRequest("POST", conf.ToolbarURL+"?"+query, nil)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		if err := resp.Body.Close(); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}

		var toolbarData map[string]interface{}
		if err := json.Unmarshal(respBody, &toolbarData); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		ctx.Set("toolbar", toolbarData)

		user, ok := toolbarData["user"].(map[string]interface{})
		if ok {
			userID, ok := user["id"]
			if ok {
				ctx.Set("authenticated", true)
				ctx.Set("userID", userID)
			}
		}

		ctx.Next()
	}
}

func AbortUnauthorized() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authenticated, exists := ctx.Get("authenticated")
		if exists && authenticated == false {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, map[string]bool{"authorized": false})
		}
		ctx.Next()
	}
}
