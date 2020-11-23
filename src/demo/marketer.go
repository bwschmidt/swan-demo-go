/* ****************************************************************************
 * Copyright 2020 51 Degrees Mobile Experts Limited (51degrees.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 * ***************************************************************************/

package demo

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"swan"
)

type marketer struct {
	config  *Configuration // Configuration information for the demo
	request *http.Request  // The background color for the page.
	results []*swan.Pair   // The results for display
	bid     string
}

func (m *marketer) Title() string { return m.request.Host }
func (m *marketer) BackgroundColor() string {
	return getBackgroundColor(m.request.Host)
}
func (m *marketer) JSON() string {
	sd, err := base64.URLEncoding.DecodeString(m.bid)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(sd)
}
func (m *marketer) Results() []*swan.Pair { return m.results }
func (m *marketer) SWANURL() string {
	u, _ := createUpdateURL(m.config, m.request)
	return u
}

func handlerMarketer(c *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m marketer

		m.config = c
		m.request = r

		err := r.ParseForm()
		if err != nil {
			returnServerError(c, w, err)
			return
		}

		bid := r.FormValue("bid")
		m.bid = bid

		err = marTemplate.Execute(w, &m)
		if err != nil {
			returnServerError(c, w, err)
		}
	}
}
