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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"swan"
)

type preference struct {
	config  *Configuration // Configuration information for the demo
	request *http.Request  // The background color for the page.
	results []*swan.Pair   // The results for display
}

func (p *preference) CBID() *swan.Pair  { return findResult(p, "cbid") }
func (p *preference) SID() *swan.Pair   { return findResult(p, "sid") }
func (p *preference) Allow() *swan.Pair { return findResult(p, "allow") }
func (p *preference) Pubs() []string    { return p.config.Pubs }
func (p *preference) Title() string     { return p.request.Host }
func (p *preference) BackgroundColor() string {
	return getBackgroundColor(p.request.Host)
}
func (p *preference) SWANURL() string {
	u, _ := createUpdateURL(p.config, p.request)
	return u
}
func (p *preference) Results() []*swan.Pair { return p.results }
func (p *preference) JSON() string {
	b, err := json.Marshal(p.results)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}
func (p *preference) IsSet() bool {
	for _, e := range p.results {
		if e.Key == "allow" {
			o, err := e.AsOWID()
			if err != nil {
				return false
			}
			if o.PayloadAsString() != "" {
				return true
			}
		}
	}
	return false
}

func handlerPublisher(c *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p *preference
		var err error

		// Only process the request if the host is on the list of publishers.
		if find(c.Pubs, r.Host) == false {
			return
		}

		// Try the URL path for the preference values.
		if r.URL.Path != "/" {
			p, err = newPreferencesFromPath(c, r)
			if err != nil {
				returnServerError(c, w, err)
				return
			}
			if p != nil {
				setCookies(r, w, p)
				http.Redirect(
					w,
					r,
					fmt.Sprintf("%s://%s%s", c.Scheme, r.Host, r.URL.Path),
					303)
			}
		}

		// If the path does not contain any values then get them from the
		// cookies.
		if p == nil {
			p = newPreferencesFromCookies(r)
		}

		if p != nil && len(p.results) > 0 {
			// Have the user provided values been set yet?
			if p.IsSet() {
				// Yes, so display the page.
				p.request = r
				p.config = c
				err = pubTemplate.Execute(w, &p)
				if err != nil {
					returnServerError(c, w, err)
				}
			} else {
				// No, so ask the user from SWAN.
				u, err := createUpdateURL(p.config, p.request)
				if err != nil {
					returnServerError(c, w, err)
				}
				http.Redirect(w, r, u, 303)
			}
		} else {
			// No preferences so start the process to fetch them from SWAN.
			u, err := createFetchURL(c, r)
			if err != nil {
				returnServerError(c, w, err)
			}
			http.Redirect(w, r, u, 303)
		}
	}
}

func newPreferencesFromCookies(r *http.Request) *preference {
	var p preference
	for _, c := range r.Cookies() {
		if c.Name == "cbid" || c.Name == "sid" || c.Name == "allow" {
			var s swan.Pair
			s.Key = c.Name
			s.Value = c.Value
			s.Expires = c.Expires
			p.results = append(p.results, &s)
		}
	}
	return &p
}

func setCookies(r *http.Request, w http.ResponseWriter, p *preference) {
	for _, i := range p.results {
		c := http.Cookie{
			Name:     i.Key,
			Domain:   getDomain(r.Host),
			Path:     r.URL.Path,
			Value:    i.Value,
			SameSite: http.SameSiteLaxMode,
			HttpOnly: true,
			Expires:  i.Expires}
		http.SetCookie(w, &c)
	}
}

func getDomain(h string) string {
	s := strings.Split(h, ":")
	return s[0]
}

func newPreferencesFromPath(
	config *Configuration,
	r *http.Request) (*preference, error) {

	var p preference // Preferences for display in the HTML

	// The last path segment is the data.
	l := strings.LastIndex(r.URL.Path, "/")
	if l < 0 {
		return nil, fmt.Errorf("URL path contains no SWAN data")
	}

	// Decrypt the string with SWAN.
	in, err := pubDecrypt(config, r, r.URL.Path[l+1:])
	if err != nil {
		return nil, err
	}

	// Get the results.
	err = json.Unmarshal(in, &p.results)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func pubDecrypt(config *Configuration, r *http.Request, q string) ([]byte, error) {

	// Combine it with the access node to decrypt the result.
	u, err := url.Parse(config.Scheme + "://" + config.SwanDomain)
	if err != nil {
		return nil, err
	}
	u.Path = "/swan/api/v1/decrypt"
	u.RawQuery = q

	// Call the URL and unpack the results if they're available.
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(u.String(), res)
	}
	return ioutil.ReadAll(res.Body)
}

func getPage(config *Configuration, r *http.Request) string {
	return fmt.Sprintf("%s://%s%s",
		config.Scheme,
		r.Host,
		r.URL.Path)
}

func setCommon(config *Configuration, r *http.Request, q *url.Values) {

	// Set the url for the return operation.
	q.Set("returnUrl", getPage(config, r))

	// Set the remote address and X-FORWARDED-FOR header.
	if r.Header.Get("X-FORWARDED-FOR") != "" {
		q.Set("X-FORWARDED-FOR", r.Header.Get("X-FORWARDED-FOR"))
	}
	q.Set("remoteAddr", r.RemoteAddr)

	// Set the user interface title, message and colours.
	q.Set("title", "SWAN demo preferences")
	q.Set("message", "Hang tight. Handling your preferences.")
	q.Set("backgroundColor", getBackgroundColor(r.Host))
	q.Set("progressColor", "darkblue")
	q.Set("messageColor", "black")
	q.Set("bounces", "10")
}

func createUpdateURL(config *Configuration, r *http.Request) (string, error) {
	return createSwanURL(config, r, "update")
}

func createFetchURL(config *Configuration, r *http.Request) (string, error) {
	return createSwanURL(config, r, "fetch")
}

func createSwanURL(
	config *Configuration,
	r *http.Request,
	action string) (string, error) {

	// Build a new URL to request the first storage operation URL.
	u, err := url.Parse(
		config.Scheme + "://" + config.SwanDomain + "/swan/api/v1/" + action)
	if err != nil {
		return "", err
	}

	// Add the query string paramters for the share web state operation.
	q := u.Query()
	setCommon(config, r, &q)
	u.RawQuery = q.Encode()

	// Get the first storage operation URL from the access node.
	res, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", newResponseError(u.String(), res)
	}

	// Read the response as a string.
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func find(a []string, v string) bool {
	for _, n := range a {
		if v == n {
			return true
		}
	}
	return false
}

func findResult(p *preference, k string) *swan.Pair {
	for _, n := range p.results {
		if k == n.Key {
			return n
		}
	}
	return nil
}
