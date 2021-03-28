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
	"encoding/json"
	"fmt"
	"fod"
	"log"
	"net/http"
	"net/url"
	"swan"
)

// Handler for publisher web pages.
func Handler(d *common.Domain, w http.ResponseWriter, r *http.Request) {

	// Check to see if this request is for an advert.
	if r.URL.Path == "/advert" {
		HandlerAdvert(d, w, r)
		return
	}

	// Try the URL path for the preference values.
	p, ae := newSWANDataFromPath(d, r)
	if ae != nil {

		// If the data can't be decrypted rather than another type of error
		// then redirect to the CMP dialog.
		if ae.StatusCode() >= 400 && ae.StatusCode() < 500 {
			if d.SwanPostMessage == false {
				http.Redirect(w, r, getCMPURL(d, r), 303)
			} else {
				handlerPublisherPage(d, w, r, p)
			}
			return
		}
		common.ReturnServerError(d.Config, w, ae)
		return
	}
	if p != nil {
		redirectToCleanURL(d.Config, w, r, p)
		return
	}

	// If the path does not contain any values then get them from the cookies.
	if p == nil {
		var err error
		p, err = newSWANDataFromCookies(r)
		if err != nil && d.Config.Debug {
			log.Println(err.Error())
		}
	}

	// If the request is from a crawler than ignore SWAN.
	c, err := fod.GetCrawlerFrom51Degrees(r)
	if err != nil {
		common.ReturnServerError(d.Config, w, err)
		return
	}
	if c {
		handlerPublisherPage(d, w, r, p)
		return
	}

	// If there is valid SWAN data then display the page using the page handler.
	// If the SWAN data is not complete or valid then ask the user to verify
	// or add the required data via the update redirect action.
	// If the SWAN data is not present or invalid then redirect to SWAN to
	// get the latest data.
	if p != nil && len(p) > 0 {
		if isSet(p) {
			handlerPublisherPage(d, w, r, p)
		} else {
			http.Redirect(w, r, getCMPURL(d, r), 303)
		}
	} else {
		if d.SwanPostMessage == false {
			redirectToSWANFetch(d, w, r)
		} else {
			handlerPublisherPage(d, w, r, p)
		}
	}
}

func handlerPublisherPage(
	d *common.Domain,
	w http.ResponseWriter,
	r *http.Request,
	p []*swan.Pair) {
	t := d.LookupHTML(r.URL.Path)
	if t == nil {
		http.NotFound(w, r)
		return
	}
	var m Model
	m.Domain = d
	m.Request = r
	m.results = p
	g := gzip.NewWriter(w)
	defer g.Close()
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	err := t.Execute(g, &m)
	if err != nil {
		common.ReturnServerError(d.Config, w, err)
	}
}

func newSWANDataFromCookies(r *http.Request) ([]*swan.Pair, error) {
	var p []*swan.Pair
	for _, c := range r.Cookies() {
		if c.Name == "swid" || c.Name == "sid" ||
			c.Name == "pref" || c.Name == "stop" {
			i, err := swan.NewPairFromCookie(c)
			if err != nil {
				return nil, err
			}
			p = append(p, i)
		}
	}
	return p, nil
}

func newSWANData(
	d *common.Domain,
	v string) ([]*swan.Pair, *common.SWANError) {
	var p []*swan.Pair

	// Decrypt the SWAN data string.
	in, e := decode(d, v)
	if e != nil {
		return nil, e
	}

	// If debug is enabled then output the JSON.
	if d.Config.Debug {
		log.Println(string(in))
	}

	// Get the results.
	err := json.Unmarshal(in, &p)
	if err != nil {
		return nil, &common.SWANError{Err: err}
	}

	return p, nil
}

func newSWANDataFromPath(
	d *common.Domain,
	r *http.Request) ([]*swan.Pair, *common.SWANError) {

	// Get the section of the URL that has the SWAN data.
	b := common.GetSWANDataFromRequest(r)
	if b == "" {
		return nil, nil
	}

	return newSWANData(d, b)
}

// SWAN data could be obtained from the URL. Remove the SWAN data string from
// the URL and redirect back to the page. Set cookies in the redirect so that
// the data is persisted.
func redirectToCleanURL(
	c *common.Configuration,
	w http.ResponseWriter,
	r *http.Request,
	p []*swan.Pair) {
	u := common.GetCleanURL(c, r).String()
	if c.Debug {
		log.Printf("Redirecting to '%s'\n", u)
	}
	setCookies(r, w, p)
	http.Redirect(w, r, u, 303)
}

func redirectToSWANFetch(
	d *common.Domain,
	w http.ResponseWriter,
	r *http.Request) {
	u, err := getSWANURL(d, r)
	if err != nil {
		common.ReturnProxyError(d.Config, w, err)
		return
	}
	http.Redirect(w, r, u, 303)
}

func getSWANURL(
	d *common.Domain,
	r *http.Request) (string, *common.SWANError) {
	return d.CreateSWANURL(
		r,
		common.GetCleanURL(d.Config, r).String(),
		"fetch",
		func(q url.Values) {
			setFlags(d, &q)
			if d.SwanNodeCount > 0 {
				q.Set("nodeCount", fmt.Sprintf("%d", d.SwanNodeCount))
			}
		})
}

func setFlags(d *common.Domain, q *url.Values) {
	if d.SwanPostMessage {
		q.Set("postMessageOnComplete", "true")
	} else {
		q.Set("postMessageOnComplete", "false")
	}
	if d.SwanDisplayUserInterface {
		q.Set("displayUserInterface", "true")
	} else {
		q.Set("displayUserInterface", "false")
	}
}

func getHomeNode(
	d *common.Domain,
	r *http.Request) (string, *common.SWANError) {
	b, err := d.CallSWANURL("home-node", nil)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func setCookies(r *http.Request, w http.ResponseWriter, p []*swan.Pair) {
	var s bool
	if r.URL.Scheme == "https" {
		s = true
	} else {
		s = false
	}
	for _, i := range p {
		http.SetCookie(w, i.AsCookie(r, w, s))
	}
}

// Returns the CMP preferences URL.
func getCMPURL(d *common.Domain, r *http.Request) string {
	var u url.URL
	u.Scheme = d.Config.Scheme
	u.Host = d.CMP
	u.Path = "/preferences"
	q := u.Query()
	q.Set("returnUrl", common.GetCleanURL(d.Config, r).String())
	q.Set("accessNode", d.SWANAccessNode)
	setFlags(d, &q)
	u.RawQuery = q.Encode()
	return u.String()
}

// isSet returns true if all three of the values are present in the results and
// are valid OWIDs.
func isSet(d []*swan.Pair) bool {
	c := 0
	for _, e := range d {
		if e.Key == "pref" || e.Key == "swid" || e.Key == "sid" {
			o, err := e.AsOWID()
			if err != nil {
				return false
			}
			if len(o.Payload) > 0 {
				c++
			}
		}
	}
	return c == 3
}

func decode(d *common.Domain, v string) ([]byte, *common.SWANError) {
	return d.CallSWANURL("data", func(q url.Values) error {
		q.Set("data", v)
		return nil
	})
}
