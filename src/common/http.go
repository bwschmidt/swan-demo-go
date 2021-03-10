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

package common

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Handler for all HTTP requests to domains controlled by the demo.
func Handler(d []*Domain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Set to true if a domain is found and handled.
		found := false

		// r.Host may include the port number or www prefixes or other
		// characters dependent on the environment. Using strings.Contains
		// rather than testing for equality eliminates these issues for a demo
		// where the domain names are not sub strings of one another.
		for _, domain := range d {
			if strings.EqualFold(r.Host, domain.Host) {

				// Try static resources first.
				f, err := handlerStatic(domain, w, r)
				if err != nil {
					ReturnServerError(domain.Config, w, err)
					return
				}

				// If not found then use the domain handler.
				if f == false {
					domain.handler(domain, w, r)
				}

				// Mark as the domain being found and then break.
				found = true
				break
			}
		}

		// All handlers have been tried and nothing has been found. Return not
		// found.
		if found == false {
			http.NotFound(w, r)
		}
	}
}

// NewResponseError used to produce an error response to a request.
func NewResponseError(c *Configuration, r *http.Response) error {
	in, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	var u string
	if c.Debug {
		u = r.Request.URL.String()
	} else {
		u = r.Request.Host
	}
	return fmt.Errorf("API call '%s' returned '%d' and '%s'",
		u,
		r.StatusCode,
		strings.TrimSpace(string(in)))
}

// ReturnServerError returns an error for a UI page request.
func ReturnServerError(c *Configuration, w http.ResponseWriter, err error) {
	if c.Debug {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if c.Debug {
		println(err.Error())
	}
}

// ReturnAPIError an API error
func ReturnAPIError(
	c *Configuration,
	w http.ResponseWriter,
	err error,
	code int) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Error(w, err.Error(), code)
	if c.Debug {
		println(err.Error())
	}
}

func getCurrentPage(c *Configuration, r *http.Request) string {
	var u url.URL
	u.Scheme = c.Scheme
	u.Host = r.Host
	u.Path = r.URL.Path
	return u.String()
}

// GetReferer returns a parsed URL from the referer header.
func GetReferer(r *http.Request) (string, error) {
	u, err := url.Parse(r.Header.Get("Referer"))
	if err != nil {
		return "", err
	}
	u.RawQuery = ""
	return u.String(), nil
}
