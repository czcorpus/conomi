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
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserID(ctx *gin.Context) (int, error) {
	ctxUserID, exists := ctx.Get("userID")
	if !exists {
		return 0, fmt.Errorf("user ID not found")
	}
	userID, ok := ctxUserID.(string)
	if !ok {
		return 0, fmt.Errorf("user ID has to be string number")
	}
	return strconv.Atoi(userID)
}
