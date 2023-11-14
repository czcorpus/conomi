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

type authArgs struct {
	SID      string `json:"sid"`
	At       string `json:"at"`
	Rmme     string `json:"rmme"`
	Lang     string `json:"lang"`
	Current  string `json:"current"`
	Continue string `json:"continue"`
}

func Auth(conf *AuthConf, publicPath string) gin.HandlerFunc {
	if conf == nil {
		return func(ctx *gin.Context) {
			ctx.Set("authenticated", true)
			ctx.Next()
		}
	}
	return func(ctx *gin.Context) {
		cookieSID, _ := ctx.Cookie(conf.CookieSID)
		cookieAt, _ := ctx.Cookie(conf.CookieAt)
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
		params.Add("current", "kontext")
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

		var data map[string]interface{}
		err = json.Unmarshal(respBody, &data)
		log.Debug().Any("resp", data).Msg("")
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		redirect, ok := data["redirect"].(string)
		if ok {
			ctx.Redirect(http.StatusMovedPermanently, redirect)
		}

		user, ok := data["user"].(map[string]interface{})
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
		_, ok = user["id"].(int)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		ctx.Next()
	}
}