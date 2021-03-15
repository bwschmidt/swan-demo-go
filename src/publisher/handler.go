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
	"fod"
	"log"
	"net/http"
	"net/url"
	"strings"
	"swan"
)

// Handler for publisher web pages.
func Handler(d *common.Domain, w http.ResponseWriter, r *http.Request) {

	// Try the URL path for the preference values.
	p, ae := newSWANDataFromPath(d, r)
	if ae != nil {

		// If the data can't be decrypted rather than another type of error
		// then redirect via SWAN to the dialog.
		if ae.StatusCode() >= 400 && ae.StatusCode() < 500 {
			redirectToSWANFetch(d, w, r)
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
		p = newSWANDataFromCookies(r)
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
		redirectToSWANFetch(d, w, r)
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

func newSWANDataFromCookies(r *http.Request) []*swan.Pair {
	var p []*swan.Pair
	for _, c := range r.Cookies() {
		if c.Name == "cbid" || c.Name == "sid" ||
			c.Name == "allow" || c.Name == "stop" {
			var s swan.Pair
			s.Key = c.Name
			s.Value = string(c.Value)
			p = append(p, &s)
		}
	}
	return p
}

func newSWANDataFromPath(
	d *common.Domain,
	r *http.Request) ([]*swan.Pair, *common.SWANError) {
	var p []*swan.Pair

	// Get the section of the URL that has the SWAN data.
	b := common.GetSWANDataFromRequest(r)
	if b == "" {
		return nil, nil
	}

	// Decrypt the SWAN data string.
	in, e := decode(d, b)
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
		return nil, &common.SWANError{err, nil}
	}

	return p, nil
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
	u, err := d.CreateSWANURL(
		r,
		common.GetCleanURL(d.Config, r).String(),
		"fetch",
		nil)
	if err != nil {
		common.ReturnProxyError(d.Config, w, err)
		return
	}
	http.Redirect(w, r, u, 303)
}

func setCookies(r *http.Request, w http.ResponseWriter, p []*swan.Pair) {
	var s bool
	if r.URL.Scheme == "https" {
		s = true
	} else {
		s = false
	}
	for _, i := range p {
		c := http.Cookie{
			Name:     i.Key,
			Domain:   getDomain(r.Host),    // Specifically to this domain
			Value:    i.Value,              // The OWID value
			SameSite: http.SameSiteLaxMode, // Available to all paths
			// The cookie never needs to be read from JavaScript so always true
			HttpOnly: true,
			Secure:   s, // Secure if HTTPs, otherwise false.
			// Set the cookie expiry time to the same as the SWAN pair.
			Expires: i.Expires,
		}
		http.SetCookie(w, &c)
	}
}

func getDomain(h string) string {
	s := strings.Split(h, ":")
	return s[0]
}

// Returns the CMP preferences URL.
func getCMPURL(d *common.Domain, r *http.Request) string {
	var u url.URL
	u.Scheme = d.Config.Scheme
	u.Host = d.CMP
	u.Path = "/preferences"
	q := u.Query()
	q.Set("returnUrl", common.GetCurrentPage(d.Config, r).String())
	q.Set("accessNode", d.SWANAccessNode)
	u.RawQuery = q.Encode()
	return u.String()
}

// isSet returns true if all three of the values are present in the results and
// are valid OWIDs.
func isSet(d []*swan.Pair) bool {
	c := 0
	for _, e := range d {
		if e.Key == "allow" || e.Key == "cbid" || e.Key == "sid" {
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
	return d.CallSWANURL("values-as-json", func(q url.Values) error {
		q.Set("data", v)
		return nil
	})
}
