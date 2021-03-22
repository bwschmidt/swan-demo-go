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

package cmp

import (
	"common"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/url"
)

// handlerStop matches the path /stop and redirects the response to the
// SWAN stop preferences URL where the parameters include the host that should
// be stopped from displaying adverts on the browser.
func handlerStop(d *common.Domain, w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		common.ReturnServerError(d.Config, w, err)
		return
	}

	if r.Form.Get("host") == "" {
		common.ReturnStatusCodeError(
			d.Config,
			w,
			fmt.Errorf("Host to be stopped must be provided"),
			http.StatusBadRequest)
		return
	}

	u, e := d.CreateSWANURL(
		r,
		r.Form.Get("returnUrl"),
		"stop",
		func(q url.Values) {
			q.Set("host", r.Form.Get("host"))

			// Demonstrate the CMP can control the nodes used.
			q.Set("nodeCount", "30")

			// Demonstrate the CMP can set all the messages.
			if q.Get("message") == "" {
				q.Set("message", fmt.Sprintf(
					"Bye, bye %s. Thanks for letting us know.",
					r.Form.Get("host")))
			}
		})
	if e != nil {
		common.ReturnProxyError(d.Config, w, e)
		return
	}

	// Return the URL as a text string.
	g := gzip.NewWriter(w)
	defer g.Close()
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, err = g.Write([]byte(u))
	if err != nil {
		common.ReturnServerError(d.Config, w, err)
		return
	}
}
