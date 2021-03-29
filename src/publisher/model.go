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
	"bytes"
	"common"
	"encoding/base64"
	"fmt"
	"html/template"
	"math/rand"
	"net/url"
	"openrtb"
	"owid"
	"strings"
	"swan"
	"time"
)

// Model used with HTML templates.
type Model struct {
	common.PageModel
	swanData []*swan.Pair // The SWAN data for display
}

// CMPURL returns the URL for the CMP dialog.
func (m Model) CMPURL() string {
	return getCMPURL(m.Domain, m.Request, m.swanData)
}

// SWANURL returns the URL for the SWAN operation. Used when the SWAN
// transaction is completed in a pop up window or iFrame.
func (m Model) SWANURL() string {
	u, _ := getSWANURL(m.Domain, m.Request, m.swanData)
	return u
}

// HomeNode returns the internet domain name of the home node.
func (m Model) HomeNode() string {
	h, _ := getHomeNode(m.Domain, m.Request)
	return h
}

// IsNew returns true if the SWID is newly created, otherwise false.
func (m Model) IsNew() bool {
	o, _ := m.swid().AsOWID()
	if o != nil {
		return o.Age() <= 1
	}
	return false
}

// Personalized returns a boolean to indicate if personalized marketing is enabled.
func (m Model) Personalized() bool { return m.PrefAsString() == "on" }

// SWIDAsString Secure Web IDentifier
func (m Model) SWIDAsString() string { return common.AsStringFromUUID(m.swid()) }

// SIDAsString Signed in IDentifier
func (m Model) SIDAsString() string { return common.AsPrintable(m.sid()) }

// PrefAsString true if personalized marketing allowed, otherwise false
func (m Model) PrefAsString() string { return common.AsString(m.pref()) }

// SWIDDomain returns the domain that created the SWID OWID
func (m Model) SWIDDomain() string { return common.OWIDDomain(m.swid()) }

// SIDDomain returns the domain that created the SID OWID
func (m Model) SIDDomain() string { return common.OWIDDomain(m.sid()) }

// PrefDomain returns the domain that created the Allow OWID
func (m Model) PrefDomain() string { return common.OWIDDomain(m.pref()) }

// SWIDDate returns the date SWID OWID was created
func (m Model) SWIDDate() string { return common.OWIDDate(m.swid()) }

// SIDDate returns the date SID OWID was created
func (m Model) SIDDate() string { return common.OWIDDate(m.sid()) }

// PrefDate returns the date Allow OWID was created
func (m Model) PrefDate() string { return common.OWIDDate(m.pref()) }

// Stopped returns a list of the domains that have been stopped for advertising.
func (m Model) Stopped() []string {
	return strings.Split(common.AsString(m.stop()), "\r\n")
}

// DomainsByCategory returns all the domains that match the category.
func (m Model) DomainsByCategory(category string) []*common.Domain {
	var domains []*common.Domain
	for _, domain := range m.Domain.Config.Domains {
		if domain.Category == category {
			domains = append(domains, domain)
		}
	}
	return domains
}

// NewAdvertHTML provides the HTML for the advert that will be displayed on the
// web page at the placement provided.
func (m Model) NewAdvertHTML(placement string) (template.HTML, error) {

	// Check that the preference information has a value and is not empty.
	if m.PrefAsString() == "" {
		return template.HTML("<p>Preferences not set</p>"), nil
	}

	rand.Seed(time.Now().UTC().UnixNano())

	// Use the SWAN network to generate the Offer ID.
	r, ae := m.newOfferID(placement)
	if ae != nil {
		return "", ae.Err
	}

	// Add the publishers signature and then process the supply chain.
	_, err := openrtb.HandleTransaction(m.Domain, r)
	if err != nil {
		return template.HTML("<p>" + err.Error() + "</p>"), nil
	}

	// Get the OWID tree as a base 64 string.
	e, err := r.AsJSON()
	if err != nil {
		return template.HTML("<p>" + err.Error() + "</p>"), nil
	}

	// Get the winning bid node.
	w, err := swan.WinningNode(r)
	if err != nil {
		return template.HTML("<p>" + err.Error() + "</p>"), nil
	}

	// Get the winning bid.
	b, err := swan.WinningBid(r)
	if err != nil {
		return template.HTML("<p>" + err.Error() + "</p>"), nil
	}

	// Get the return URL.
	t, err := common.GetReturnURL(m.Request)
	if err != nil {
		return template.HTML("<p>" + err.Error() + "</p>"), nil
	}

	// Get the URL for the info icon.
	var i url.URL
	i.Scheme = m.Config().Scheme
	i.Host = m.Domain.CMP
	i.Path = "/info"
	q := i.Query()
	n := w
	for n != nil {
		q.Add("owid", n.GetOWIDAsString())
		n = n.GetParent()
	}
	q.Set("returnUrl", t.String())
	i.RawQuery = q.Encode()

	// Return a FORM HTML element with a button for the advert. The OWID tree
	// is a base 64 string added as a hidden field to the form.
	var html bytes.Buffer
	html.WriteString(fmt.Sprintf("<form method=\"POST\" action=\"//%s\">"+
		"<div class=\"form-group\">"+
		"<input type=\"hidden\" id=\"transaction\" name=\"transaction\" value=\"%s\">"+
		"<button type=\"submit\" id=\"view\" name=\"view\" class=\"advert-button\">"+
		"<img src=\"//%s\">"+
		"</button>"+
		"<a href=\"%s\" class=\"advert-stop\" title=\"Info about this advert\">"+
		"<img src=\"%s\">"+
		"</a>"+
		"</div>"+
		"</form>",
		b.AdvertiserURL,
		base64.RawStdEncoding.EncodeToString(e),
		b.MediaURL,
		i.String(),
		"noun_Info_1582932.svg"))
	return template.HTML(html.String()), nil
}

// SWID Secure Web IDentifier
func (m Model) swid() *swan.Pair { return m.findResult("swid") }

// SID Signed in IDentifier
func (m Model) sid() *swan.Pair { return m.findResult("sid") }

// Allow true if personalized marketing allowed, otherwise false
func (m Model) pref() *swan.Pair { return m.findResult("pref") }

// Stop the list of domains that should not have adverts displayed form.
func (m Model) stop() *swan.Pair { return m.findResult("stop") }

func (m Model) findResult(k string) *swan.Pair {
	for _, n := range m.swanData {
		if strings.EqualFold(k, n.Key) {
			return n
		}
	}
	return nil
}

// newOfferID returns a new Offer OWID Node from the SWAN network.
func (m *Model) newOfferID(placement string) (*owid.Node, *common.SWANError) {
	var n owid.Node
	var err *common.SWANError
	if m.swid() == nil {
		return nil, &common.SWANError{Err: fmt.Errorf("SWID missing")}
	}
	if m.sid() == nil {
		return nil, &common.SWANError{Err: fmt.Errorf("SID missing")}
	}
	if m.pref() == nil {
		return nil, &common.SWANError{Err: fmt.Errorf("Pref missing")}
	}
	if m.stop() == nil {
		return nil, &common.SWANError{Err: fmt.Errorf("Stop missing")}
	}
	n.OWID, err = m.Domain.CallSWANURL("create-offer-id",
		func(q url.Values) error {
			q.Add("placement", placement)
			q.Add("pubdomain", m.Request.Host)
			q.Add("swid", m.swid().Value)
			q.Add("sid", m.sid().Value)
			q.Add("pref", m.pref().Value)
			q.Add("stop", m.stop().Value)
			return nil
		})
	if err != nil {
		return nil, err
	}
	return &n, nil
}
