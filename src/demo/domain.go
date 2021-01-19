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
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"owid"
	"path/filepath"
)

// Domain represents the information held in the domain configuration file
// commonly represented in the demo in config.json.
type Domain struct {
	Category            string             // Category of the domain
	Name                string             // Common name for the domain
	Bad                 bool               // True if this domain is a bad actor for the demo
	Host                string             // The host name for the domain
	SwanMessage         string             // Message if used with SWAN
	SwanBackgroundColor string             // Background color if used with SWAN
	SwanMessageColor    string             // Message text color if used with SWAN
	SwanProgressColor   string             // Message progress color if used with SWAN
	SwanHost            string             // Access node domain for if used with SWAN
	Suppliers           []string           // Suppliers used by the domain operator
	Adverts             []Advert           // Adverts the domain can serve
	config              *Configuration     // Configuration for the server
	folder              string             // Location of the directory
	templates           *template.Template // HTML templates
	owid                *owid.Creator      // The OWID creator associated with the domain if any
}

// newDomain creates a new instance of domain information from the file
// provided.
func newDomain(c *Configuration, folder string) (*Domain, error) {
	var d Domain

	// Read the configuration for the folder provided.
	configFile, err := os.Open(filepath.Join(folder, "config.json"))
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&d)

	// Set the private members.
	d.config = c
	d.Host = filepath.Base(folder)
	d.folder = folder
	d.templates, err = d.parseHTML()
	if err != nil {
		return nil, err
	}
	d.owid, err = c.owid.GetCreator(d.Host)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (d *Domain) parseHTML() (*template.Template, error) {
	var htmlFiles []string
	files, err := ioutil.ReadDir(d.folder)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".html" {
			htmlFiles = append(
				htmlFiles,
				filepath.Join(d.folder, file.Name()))
		}
	}
	if len(htmlFiles) > 0 {
		return template.ParseFiles(htmlFiles...)
	}
	return nil, nil
}

func (d *Domain) lookupHTML(p string) *template.Template {
	if d.templates == nil {
		return nil
	}
	t := d.templates.Lookup(filepath.Base(p))
	if t == nil {
		t = d.templates.Lookup("default.html")
	}
	return t
}

func (d *Domain) setCommon(r *http.Request, q *url.Values) {

	// Set the access key
	q.Set("accessKey", d.config.AccessKey)

	// Set the remote address and X-FORWARDED-FOR header.
	if r.Header.Get("X-FORWARDED-FOR") != "" {
		q.Set("X-FORWARDED-FOR", r.Header.Get("X-FORWARDED-FOR"))
	}
	q.Set("remoteAddr", r.RemoteAddr)

	// Set the user interface title, message and colours.
	q.Set("title", "SWAN demo preferences")
}

func (d *Domain) createSWANActionURL(
	r *http.Request,
	returnURL string,
	action string) (string, error) {

	// Build a new URL to request the first storage operation URL.
	u, err := url.Parse(
		d.config.Scheme + "://" + d.SwanHost + "/swan/api/v1/" + action)
	if err != nil {
		return "", err
	}

	// Add the query string paramters for the SWAN operation starting with the
	// common ones that are the same for every request from this demo.
	q := u.Query()
	d.setCommon(r, &q)

	// If an explicit return URL was provided then use that. Otherwise use the
	// page for the current request.
	if returnURL != "" {
		q.Set("returnUrl", returnURL)
	} else {
		q.Set("returnUrl", getCurrentPage(d.config, r))
	}

	// Add user interface parameters for the SWAN operation and the user
	// interface.
	if d.SwanMessage != "" {
		q.Set("message", d.SwanMessage)
	}
	if d.SwanBackgroundColor != "" {
		q.Set("backgroundColor", d.SwanBackgroundColor)
	}
	if d.SwanProgressColor != "" {
		q.Set("progressColor", d.SwanProgressColor)
	}
	if d.SwanMessageColor != "" {
		q.Set("messageColor", d.SwanMessageColor)
	}
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
