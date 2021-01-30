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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"owid"
	"path/filepath"
	"swan"
	"swift"
)

// AddHandlers and outputs configuration information.
func AddHandlers(settingsFile string) {

	// Get the demo configuration.
	dc := newConfig(settingsFile)

	// Get the example simple access control implementations.
	swi := swift.NewAccessSimple(dc.AccessKeys)
	oa := owid.NewAccessSimple(dc.AccessKeys)
	swa := swan.NewAccessSimple(dc.AccessKeys)

	// Get all the domains for the SWAN demo.
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	domains, err := parseDomains(&dc, filepath.Join(wd, "www"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	dc.domains = domains

	// Add the SWAN handlers, with the demo handler being used for any
	// malformed storage requests.
	swan.AddHandlers(settingsFile, swa, swi, oa, handler(domains))

	// Output details for information.
	log.Printf("Demo scheme: %s\n", dc.Scheme)
	for _, d := range domains {
		log.Printf("%s:%s:%s", d.Category, d.Host, d.Name)
	}
}

// parseDomains returns an array of domains (e.g. swan-demo.uk) with all the
// information needed to server static, API and HTML requests.
// c is the general server configuration.
// path provides the root folder where the child folders are the names of the
// domains that the demo responds to.
func parseDomains(c *Configuration, path string) ([]*Domain, error) {
	var domains []*Domain
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range files {

		// Domains are the directories of the folder provided.
		if f.IsDir() {
			domain, err := newDomain(c, filepath.Join(path, f.Name()))
			if err != nil {
				return nil, err
			}
			domains = append(domains, domain)
		}
	}
	return domains, nil
}

// handler for all HTTP requests to domains controlled by the demo.
func handler(d []*Domain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// r.Host may include the port number or www prefixes or other
		// charaters dependent on the environment. Using strings.Contains
		// rather than testing for equality eliminates these issues for a demo
		// where the domain names are not sub strings of one another.
		for _, domain := range d {
			if r.Host == domain.Host {
				handlerDomain(domain, w, r)
				break
			}
		}
	}
}

func handlerDomain(d *Domain, w http.ResponseWriter, r *http.Request) {

	// Is this a request for a static resource?
	found, err := handleStatic(d, w, r)
	if err != nil {
		returnServerError(d.config, w, err)
		return
	}

	// Is this a request for the privacy preference updates?
	if found == false {
		found, err = handlePrivacy(d, w, r)
		if err != nil {
			returnServerError(d.config, w, err)
			return
		}
	}

	// Is this a request for an API transaction?
	if found == false {
		found, err = handleTransaction(d, w, r)
		if err != nil {
			fmt.Println(err.Error())
			returnServerError(d.config, w, err)
			return
		}
	}

	// If no static content is found then response with HTML.
	if found == false {
		handleHTML(d, w, r)
	}
}

func newResponseError(url string, resp *http.Response) error {
	in, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return fmt.Errorf("API call '%s' returned '%d' and '%s'",
		url, resp.StatusCode, in)
}

func returnServerError(c *Configuration, w http.ResponseWriter, err error) {
	if c.Debug {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if c.Debug {
		println(err.Error())
	}
}

func returnAPIError(
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
	return fmt.Sprintf("%s://%s%s", c.Scheme, r.Host, r.URL.Path)
}
