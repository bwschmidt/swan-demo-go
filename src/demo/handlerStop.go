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
	"fmt"
	"net/http"
	"net/url"
)

// handleStop matches the path /stop and redirects the response to the
// SWAN stop preferences URL where the parameters include the host that should
// be stopped from displaying adverts on the browser.
func handleStop(
	d *Domain,
	w http.ResponseWriter,
	r *http.Request) (bool, error) {
	if r.URL.Path == "/stop" {
		err := r.ParseForm()
		if err != nil {
			return true, err
		}
		if r.Form.Get("host") != "" {
			u, err := d.createSWANActionURL(
				r,
				getReferer(r),
				"stop",
				func(q *url.Values) {
					q.Set("host", r.Form.Get("host"))
					// Demonstrate the publisher can control the nodes used.
					q.Set("bounces", "30")
					// Demonstrate the publisher can set all messages.
					q.Set("message", fmt.Sprintf(
						"Bye, bye %s. Thanks for letting us know.",
						r.Form.Get("host")))
				})
			if err != nil {
				return true, err
			}
			http.Redirect(w, r, u, 303)
			return true, nil
		}
	}
	return false, nil
}
