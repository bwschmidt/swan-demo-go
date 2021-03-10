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
	"fod"
	"net/http"
	"net/url"
	"swan"
)

// PageModel used as the base for models used with HTML templates.
type PageModel struct {
	Domain *Domain // The domain associated with the request
	// The request that relates to the page request with the ParseForm method complete
	Request *http.Request
}

// PreferencesDialogURL returns the URL to display the preferences dialog.
func (m PageModel) PreferencesDialogURL() (string, error) {
	var u url.URL
	u.Scheme = m.Domain.Config.Scheme
	u.Host = m.Domain.CMP
	u.Path = "/preferences"
	return u.String(), nil
}

// IsCrawler returns true if the browser is a crawler, otherwise false.
func (m PageModel) IsCrawler() (bool, error) {
	return fod.GetCrawlerFrom51Degrees(m.Request)
}

// Config returns the domain configuration.
func (m PageModel) Config() *Configuration { return m.Domain.Config }

// OWIDDate returns the creator domain of the ID.
func OWIDDate(p *swan.Pair) string {
	if p == nil {
		return ""
	}
	o, err := p.AsOWID()
	if err != nil || o == nil {
		return ""
	}
	return o.Date.Format("2006-01-02")
}

// OWIDDomain returns the creator domain of the ID.
func OWIDDomain(p *swan.Pair) string {
	if p == nil {
		return ""
	}
	o, err := p.AsOWID()
	if err != nil || o == nil {
		return ""
	}
	return o.Domain
}

// AsString gets the value of the pair as string for display.
func AsString(p *swan.Pair) string {
	if p == nil {
		return ""
	}
	o, err := p.AsOWID()
	if err != nil || o == nil {
		return ""
	}
	return o.PayloadAsString()
}

// AsPrintable gets the value of the pair as a printable string for display.
func AsPrintable(p *swan.Pair) string {
	if p == nil {
		return ""
	}
	o, err := p.AsOWID()
	if err != nil || o == nil {
		return ""
	}
	return o.PayloadAsPrintable()
}
