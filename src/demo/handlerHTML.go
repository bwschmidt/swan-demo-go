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
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"swan"
)

func handleHTML(d *Domain, w http.ResponseWriter, r *http.Request) {

	// Parse the incoming parameters.
	err := r.ParseForm()
	if err != nil {
		returnServerError(d.config, w, err)
		return
	}

	// If this is a domain that support SWAN then process with SWAN, otherwise
	// jump straight to the non SWAN page response.
	if d.SwanHost != "" {
		handleSWAN(d, w, r)
	} else {
		handlePage(d, w, r, nil)
	}
}

func handlePage(
	d *Domain,
	w http.ResponseWriter,
	r *http.Request,
	p []*swan.Pair) {
	t := d.lookupHTML(r.URL.Path)
	if t == nil {
		http.NotFound(w, r)
		return
	}
	err := t.Execute(w, &PageModel{d, w, r, p, nil})
	if err != nil {
		returnServerError(d.config, w, err)
	}
}

func handleSWAN(d *Domain, w http.ResponseWriter, r *http.Request) {
	var err error
	var p []*swan.Pair // Key value pairs of SWAN data

	// Try the URL path for the preference values.
	if r.URL.Path != "/" {
		p, err = newSWANDataFromPath(d, r)
		if err != nil {
			returnServerError(d.config, w, err)
			return
		}
		if p != nil {
			redirectToCleanURL(d.config, w, r, p)
			return
		}
	}

	// If the path does not contain any values then get them from the
	// cookies.
	if p == nil {
		p = newSWANDataFromCookies(r)
	}

	// If the request is from a crawler than ignore SWAN.
	c, err := getDeviceFrom51Degrees(r)
	if err != nil {
		returnServerError(d.config, w, err)
		return
	}
	if c {
		handlePage(d, w, r, p)
		return
	}

	// If there is valid SWAN data then display the page using the page handler.
	// If the SWAN data is not complete or valid then ask the user to verify
	// or add the required data via the update redirect action.
	if p != nil && len(p) > 0 {
		if IsSet(p) {
			handlePage(d, w, r, p)
			return
		}
		redirectToSWAN(d, w, r, "update")
		return
	}

	// If the SWAN data is not present or invalid then redirect to SWAN to
	// get the latest data.
	redirectToSWAN(d, w, r, "fetch")
}

// SWAN data could be obtained from the URL. Remove the SWAN data string from
// the URL and redirect back to the page. Set cookies in the redirect so that
// the data is persisted.
func redirectToCleanURL(
	c *Configuration,
	w http.ResponseWriter,
	r *http.Request,
	p []*swan.Pair) {
	n := c.Scheme + "://" + r.Host + strings.ReplaceAll(
		r.URL.Path, getSWANString(r), "")
	if c.Debug {
		log.Printf("Redirecting to '%s'\n", n)
	}
	setCookies(r, w, p)
	http.Redirect(w, r, n, 303)
}

func redirectToSWAN(
	d *Domain,
	w http.ResponseWriter,
	r *http.Request,
	action string) {
	u, err := d.createSWANActionURL(r, "", action, nil)
	if err != nil {
		returnServerError(d.config, w, err)
		return
	}
	http.Redirect(w, r, u, 303)
}

func find(a []string, v string) bool {
	for _, n := range a {
		if v == n {
			return true
		}
	}
	return false
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

func newSWANDataFromPath(d *Domain, r *http.Request) ([]*swan.Pair, error) {
	var p []*swan.Pair

	// Decrypt the SWAN data string.
	in, err := decode(d, getSWANString(r))
	if err != nil {
		return nil, err
	}

	// If debug is enabled then output the JSON.
	if d.config.Debug {
		log.Println(string(in))
	}

	// Get the results.
	err = json.Unmarshal(in, &p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// IsSet returns true if all three of the values are present in the results and
// are valid OWIDs.
func IsSet(d []*swan.Pair) bool {
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

func getSWANString(r *http.Request) string {
	i := strings.LastIndex(r.URL.Path, "/")
	if i >= 0 {
		return r.URL.Path[i+1:]
	}
	return r.URL.RawQuery
}

func decode(d *Domain, v string) ([]byte, error) {

	// Combine it with the access node to decrypt the result.
	u, err := url.Parse(d.config.Scheme + "://" + d.SwanHost)
	if err != nil {
		return nil, err
	}
	u.Path = "/swan/api/v1/decode-as-json"
	q := u.Query()
	q.Set("data", v)
	q.Set("accessKey", d.config.AccessKey)
	u.RawQuery = q.Encode()

	// Call the URL and unpack the results if they're available.
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(d.config, res)
	}
	return ioutil.ReadAll(res.Body)
}

// The expires member of the cookie is not set so that it becomes a session
// cookie. This will ensure that the value is fetched from SWAN after the
// session expires. The value might have changed if the user visits another web
// site and changes their preferences.
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
