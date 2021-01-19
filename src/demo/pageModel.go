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
	"net/http"
	"owid"
	"swan"
)

// PageModel used with HTML templates.
type PageModel struct {
	Domain  *Domain             // The domain associated with the request
	writer  http.ResponseWriter // The writer for the response
	request *http.Request       // The request that relates to the page request
	results []*swan.Pair        // The SWAN data for display
	offer   *owid.OWID          // The offer and tree associated with the page
}

// CBIDAsString Common Browser IDentifier
func (m PageModel) CBIDAsString() string { return asString(m.cbid()) }

// SIDAsString Signed in IDentifier
func (m PageModel) SIDAsString() string { return asPrintable(m.sid()) }

// AllowAsString true if personalized marketing allowed, otherwise false
func (m PageModel) AllowAsString() string { return asString(m.allow()) }

// Allow returns a boolean to indicate if personalized marketing is enabled.
func (m PageModel) Allow() bool { return m.AllowAsString() == "on" }

// CBID Common Browser IDentifier
func (m PageModel) cbid() *swan.Pair { return m.findResult("cbid") }

// SID Signed in IDentifier
func (m PageModel) sid() *swan.Pair { return m.findResult("sid") }

// Allow true if personalized marketing allowed, otherwise false
func (m PageModel) allow() *swan.Pair { return m.findResult("allow") }

// Gets the value of the pair as string for display.
func asString(p *swan.Pair) string {
	if p == nil {
		return ""
	}
	o, err := p.AsOWID()
	if err != nil || o == nil {
		return ""
	}
	return o.PayloadAsString()
}

func asPrintable(p *swan.Pair) string {
	if p == nil {
		return ""
	}
	o, err := p.AsOWID()
	if err != nil || o == nil {
		return ""
	}
	return o.PayloadAsPrintable()
}

// DomainsByCategory returns all the domains that match the category.
func (m PageModel) DomainsByCategory(category string) []*Domain {
	var domains []*Domain
	for _, domain := range m.Domain.config.domains {
		if domain.Category == category {
			domains = append(domains, domain)
		}
	}
	return domains
}

// Config returns the domain configuration.
func (m *PageModel) Config() *Configuration { return m.Domain.config }

// WinningBid gets the winning bid from the winner's Processor OWID.
func (m *PageModel) WinningBid() (*swan.Bid, error) {
	w, err := m.Winner()
	if err != nil {
		return nil, err
	}
	return swan.BidFromOWID(w)
}

// Winner gets the winning Processor OWID for the transaction.
func (m *PageModel) Winner() (*owid.OWID, error) {
	w := m.offer.Find(func(n *owid.OWID) bool {
		return len(n.Payload) == 4
	})
	if w != nil {
		for len(w.Payload) == 4 {
			i := readUint32(w.Payload)
			if i >= uint32(len(w.Children)) {
				return nil, fmt.Errorf("Index '%d' out of range", i)
			}
			w = w.Children[i]
		}
	}
	return w, nil
}

func (m PageModel) findResult(k string) *swan.Pair {
	for _, n := range m.results {
		if k == n.Key {
			return n
		}
	}
	return nil
}
