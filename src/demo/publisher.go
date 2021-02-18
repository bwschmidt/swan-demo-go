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
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"owid"
	"time"
)

// NewAdvertHTML provides the HTML for the advert that will be displayed on the
// web page at the placement provided.
func (m *PageModel) NewAdvertHTML(placement string) (template.HTML, error) {
	var err error

	rand.Seed(time.Now().UTC().UnixNano())

	// Use the SWAN network to generate the Offer ID.
	m.offer, err = m.newOfferID(placement)
	if err != nil {
		return "", err
	}

	// Add the publishers signature and then process the supply chain.
	_, err = handleBid(m.Domain, m.offer)
	if err != nil {
		return template.HTML("<p>" + err.Error() + "</p>"), nil
	}

	// Get the OWID tree as a base 64 string.
	e, err := m.offer.AsJSON()
	if err != nil {
		return template.HTML("<p>" + err.Error() + "</p>"), nil
	}

	// Get the winning bid node.
	w, err := m.WinningNode()
	if err != nil {
		return template.HTML("<p>" + err.Error() + "</p>"), nil
	}

	// Get the winning bid.
	b, err := m.WinningBid()
	if err != nil {
		return template.HTML("<p>" + err.Error() + "</p>"), nil
	}

	// Get the URL for the info icon.
	var i url.URL
	i.Scheme = m.Config().Scheme
	i.Host = m.Domain.SwanHost
	i.Path = "/swan/info"
	q := i.Query()
	o := w
	for o != nil {
		q.Add("owid", o.GetOWIDAsString())
		o = o.GetParent()
	}
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

// newOfferID returns a new Offer OWID Node from the SWAN network.
func (m *PageModel) newOfferID(placement string) (*owid.Node, error) {

	u, err := url.Parse(
		m.Config().Scheme + "://" + m.Domain.SwanHost +
			"/swan/api/v1/create-offer-id")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Add("accessKey", m.Config().AccessKey)
	q.Add("placement", placement)
	q.Add("pubdomain", m.request.Host)
	cbid, err := m.cbid().AsBase64()
	if err != nil {
		return nil, err
	}
	q.Add("cbid", cbid)
	sid, err := m.sid().AsBase64()
	if err != nil {
		return nil, err
	}
	q.Add("sid", sid)
	allow, err := m.allow().AsBase64()
	if err != nil {
		return nil, err
	}
	q.Add("preferences", allow)
	stopped, err := m.stopped().AsBase64()
	if err != nil {
		return nil, err
	}
	q.Add("stopped", stopped)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code '%d' returned", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var n owid.Node
	n.OWID = body
	return &n, nil
}
