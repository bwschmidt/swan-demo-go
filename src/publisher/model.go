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
	results []*swan.Pair // The SWAN data for display
}

// CMPURL returns the URL for the CMP dialog.
func (m Model) CMPURL() string {
	return getCMPURL(m.Domain, m.Request)
}

// Allow returns a boolean to indicate if personalized marketing is enabled.
func (m Model) Allow() bool { return m.AllowAsString() == "on" }

// CBIDAsString Common Browser IDentifier
func (m Model) CBIDAsString() string { return common.AsString(m.cbid()) }

// SIDAsString Signed in IDentifier
func (m Model) SIDAsString() string { return common.AsPrintable(m.sid()) }

// AllowAsString true if personalized marketing allowed, otherwise false
func (m Model) AllowAsString() string { return common.AsString(m.allow()) }

// CBIDDomain returns the domain that created the CBID OWID
func (m Model) CBIDDomain() string { return common.OWIDDomain(m.cbid()) }

// SIDDomain returns the domain that created the SID OWID
func (m Model) SIDDomain() string { return common.OWIDDomain(m.sid()) }

// AllowDomain returns the domain that created the Allow OWID
func (m Model) AllowDomain() string { return common.OWIDDomain(m.allow()) }

// CBIDDate returns the date CBID OWID was created
func (m Model) CBIDDate() string { return common.OWIDDate(m.cbid()) }

// SIDDate returns the date SID OWID was created
func (m Model) SIDDate() string { return common.OWIDDate(m.sid()) }

// AllowDate returns the date Allow OWID was created
func (m Model) AllowDate() string { return common.OWIDDate(m.allow()) }

// Stopped returns a list of the domains that have been stopped for advertising.
func (m Model) Stopped() []string {
	return strings.Split(common.AsString(m.stopped()), "\r\n")
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
		base64.StdEncoding.EncodeToString(e),
		b.MediaURL,
		i.String(),
		"noun_Info_1582932.svg"))
	return template.HTML(html.String()), nil
}

// CBID Common Browser IDentifier
func (m Model) cbid() *swan.Pair { return m.findResult("cbid") }

// SID Signed in IDentifier
func (m Model) sid() *swan.Pair { return m.findResult("sid") }

// Allow true if personalized marketing allowed, otherwise false
func (m Model) allow() *swan.Pair { return m.findResult("allow") }

// Stop the list of domains that should not have adverts displayed form.
func (m Model) stopped() *swan.Pair { return m.findResult("stop") }

func (m Model) findResult(k string) *swan.Pair {
	for _, n := range m.results {
		if k == n.Key {
			return n
		}
	}
	return nil
}

// newOfferID returns a new Offer OWID Node from the SWAN network.
func (m *Model) newOfferID(placement string) (*owid.Node, *common.SWANError) {
	var n owid.Node
	var err *common.SWANError
	n.OWID, err = m.Domain.CallSWANURL("create-offer-id",
		func(q url.Values) error {
			q.Add("placement", placement)
			q.Add("pubdomain", m.Request.Host)
			cbid, err := m.cbid().AsBase64()
			if err != nil {
				return err
			}
			q.Add("cbid", cbid)
			sid, err := m.sid().AsBase64()
			if err != nil {
				return err
			}
			q.Add("sid", sid)
			allow, err := m.allow().AsBase64()
			if err != nil {
				return err
			}
			q.Add("preferences", allow)
			stopped, err := m.stopped().AsBase64()
			if err != nil {
				return err
			}
			q.Add("stopped", stopped)
			return nil
		})
	if err != nil {
		return nil, err
	}
	return &n, nil
}
