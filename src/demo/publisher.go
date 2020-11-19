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
	"log"
	"net/http"
	"net/url"
	"owid"
	"strings"
	"swan"
)

type preference struct {
	config  *Configuration // Configuration information for the demo
	request *http.Request  // The background color for the page.
	results []*swan.Pair   // The results for display
	offerID string         // The Offer ID for display
}

func (p *preference) CBID() *swan.Pair      { return findResult(p, "cbid") }
func (p *preference) SID() *swan.Pair       { return findResult(p, "sid") }
func (p *preference) Allow() *swan.Pair     { return findResult(p, "allow") }
func (p *preference) OID() string           { return p.offerID }
func (p *preference) Pubs() []string        { return p.config.Pubs }
func (p *preference) Title() string         { return p.request.Host }
func (p *preference) Results() []*swan.Pair { return p.results }
func (p *preference) BackgroundColor() string {
	return getBackgroundColor(p.request.Host)
}

func (p *preference) NewOfferID(placement string) string {
	oid, _ := p.createOfferID(placement)
	return oid
}

func (p *preference) UnpackOID() string {
	var o swan.OfferID
	ow, _ := owid.DecodeFromBase64(p.offerID)
	o.SetFromByteArray(ow.Payload)

	b, _ := json.Marshal(o)
	return string(b)
}

func (p *preference) SWANURL() string {
	u, _ := createUpdateURL(p.config, p.request)
	return u
}

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

				// Preferences could be obtained from the URL. Remove the SWAN
				// data string and redirect back to the page setting cookies.
				n := c.Scheme + "://" + r.Host +
					strings.ReplaceAll(r.URL.Path, getSWANString(r), "")
				if c.Debug {
					log.Printf("Redirecting to '%s'\n", n)
				}

				// Set the preferences as cookies and redirect back to the page
				// to remove the SWAN data string from the address bar.
				setCookies(r, w, p)
				http.Redirect(w, r, n, 303)
				return
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
				p.offerID = p.NewOfferID("1")
				err = pubTemplate.Execute(w, &p)
				if err != nil {
					returnServerError(c, w, err)
				}
			} else {
				// No, so ask the user from SWAN.
				u, err := createUpdateURL(c, r)
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

func getSWANString(r *http.Request) string {
	i := strings.LastIndex(r.URL.Path, "/")
	if i >= 0 {
		return r.URL.Path[i+1:]
	}
	return r.URL.RawQuery
}

func newPreferencesFromPath(
	config *Configuration,
	r *http.Request) (*preference, error) {

	var p preference // Preferences for display in the HTML

	// Decrypt the SWAN data string.
	in, err := decode(config, r, getSWANString(r))
	if err != nil {
		return nil, err
	}

	// If debug is enabled then output the JSON.
	if config.Debug {
		log.Println(string(in))
	}

	// Get the results.
	err = json.Unmarshal(in, &p.results)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func decode(config *Configuration, r *http.Request, d string) ([]byte, error) {

	// Combine it with the access node to decrypt the result.
	u, err := url.Parse(config.Scheme + "://" + config.SwanDomain)
	if err != nil {
		return nil, err
	}
	u.Path = "/swan/api/v1/decode-as-json"
	q := u.Query()
	q.Set("data", d)
	q.Set("accessKey", config.AccessKey)
	u.RawQuery = q.Encode()

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

	// Set the access key
	q.Set("accessKey", config.AccessKey)

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
}

func createUpdateURL(config *Configuration, r *http.Request) (string, error) {
	return createSWANActionURL(config, r, "update")
}

func createFetchURL(config *Configuration, r *http.Request) (string, error) {
	return createSWANActionURL(config, r, "fetch")
}

func createSWANActionURL(
	config *Configuration,
	r *http.Request,
	action string) (string, error) {

	// Build a new URL to request the first storage operation URL.
	u, err := url.Parse(
		config.Scheme + "://" + config.SwanDomain + "/swan/api/v1/" + action)
	if err != nil {
		return "", err
	}

	// Add the query string paramters for the SWIFT operation.
	q := u.Query()
	setCommon(config, r, &q)
	u.RawQuery = q.Encode()

	// Get the link to SWAN to use for the fetch operation.
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

func (p *preference) createOfferID(placement string) (string, error) {

	u, err := url.Parse(
		p.config.Scheme + "://" + p.config.SwanDomain +
			"/swan/api/v1/create-offer-id")
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("accessKey", p.config.AccessKey)
	q.Add("placement", placement)
	q.Add("pubdomain", p.request.Host)
	q.Add("cbid", p.CBID().Value)
	q.Add("sid", p.SID().Value)
	q.Add("preferences", p.Allow().Value)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
