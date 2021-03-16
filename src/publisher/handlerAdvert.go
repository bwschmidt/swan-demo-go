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

package publisher

import (
	"common"
	"compress/gzip"
	"fmt"
	"net/http"
)

// HandlerAdvert for the request for adverts for the publisher web pages.
func HandlerAdvert(d *common.Domain, w http.ResponseWriter, r *http.Request) {

	// Get the SWAN data from the request cookies.
	p := newSWANDataFromCookies(r)
	if p == nil {
		common.ReturnStatusCodeError(
			d.Config,
			w,
			fmt.Errorf("SWAN data cookies missing for request"),
			http.StatusBadRequest)
		return
	}

	// Create the model for publishers.
	var m Model
	m.Domain = d
	m.Request = r
	m.results = p

	// Get the form parameters which will include the placement.
	err := r.ParseForm()
	if err != nil {
		common.ReturnServerError(d.Config, w, err)
		return
	}

	// Use the new advert HTML to request the advert.
	t, err := m.NewAdvertHTML(r.Form.Get("placement"))
	if err != nil {
		common.ReturnServerError(d.Config, w, err)
		return
	}

	// Respond with the HTML.
	g := gzip.NewWriter(w)
	defer g.Close()
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, err = g.Write([]byte(t))
	if err != nil {
		common.ReturnServerError(d.Config, w, err)
		return
	}
}
