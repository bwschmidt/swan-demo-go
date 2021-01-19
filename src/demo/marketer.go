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
	"fmt"
	"html/template"
	"owid"
	"swan"
)

// OfferID returns the offer ID string for the advert.
func (m *PageModel) OfferID() (string, error) {
	t, err := m.getTransaction()
	if err != nil {
		return "", err
	}
	if t == nil {
		return "", nil
	}
	return t.TreeAsBase64()
}

// OfferIDUnpacked returns the unpacked Offer ID
func (m *PageModel) OfferIDUnpacked() (template.HTML, error) {
	t, err := m.getTransaction()
	if err != nil {
		return "", err
	}
	if t == nil {
		return template.HTML("<p>Advert not source of request.</p>"), nil
	}
	s, err := swan.OfferFromOWID(t)
	if err != nil {
		return "", err
	}

	var html bytes.Buffer
	html.WriteString("<table class=\"table\">")
	html.WriteString("<thead><tr>")
	html.WriteString("<th>Field</th>")
	html.WriteString("<th>Value</th>")
	html.WriteString("</tr></thead><tbody>")
	html.WriteString(fmt.Sprintf(
		"<tr><td>Version</td><td>%d</td></tr>",
		t.Version))
	html.WriteString(fmt.Sprintf(
		"<tr><td>Domain</td><td>%s</td></tr>",
		t.Domain))
	html.WriteString(fmt.Sprintf(
		"<tr><td>Signature</td><td style=\"word-break:break-all\">%s</td></tr>",
		convertToString(t.Signature)))
	html.WriteString(fmt.Sprintf(
		"<tr><td>CBID</td><td>%s</td></tr>",
		s.CBID))
	html.WriteString(fmt.Sprintf(
		"<tr><td>Allow</td><td>%s</td></tr>",
		s.Preferences))
	html.WriteString(fmt.Sprintf(
		"<tr><td>SID</td><td>%s</td></tr>",
		s.SID))
	html.WriteString(fmt.Sprintf(
		"<tr><td>Pub. domain</td><td>%s</td></tr>",
		s.PubDomain))
	html.WriteString(fmt.Sprintf(
		"<tr><td>Placement</td><td>%s</td></tr>",
		s.Placement))
	html.WriteString(fmt.Sprintf(
		"<tr><td>Unique</td><td style=\"word-break:break-all\">%s</td></tr>",
		convertToString(s.UUID)))
	htmlAddFooter(&html)
	return template.HTML(html.String()), nil
}

func convertToString(b []byte) string {
	return fmt.Sprintf("%x", b)
}

// AuditWinnerHTML returns the audit information from the bid used in the advert
// that resulted in the request to this page.
func (m *PageModel) AuditWinnerHTML() (template.HTML, error) {
	var html bytes.Buffer
	t, err := m.getTransaction()
	if err != nil {
		if m.Config().Debug {
			html.WriteString(fmt.Sprintf(
				"<p class=\"warning\">%s</p>",
				err.Error()))
		}
		html.WriteString("<p>Advert not source of request.</p>")
		return template.HTML(html.String()), nil
	}
	if t == nil {
		return template.HTML("<p>Advert not source of request.</p>"), nil
	}

	w, err := m.Winner()
	if err != nil {
		return "", nil
	}

	if err != nil {
		return "", nil
	}
	htmlAddHeader(&html)
	err = appendParents(&html, w)
	if err != nil {
		return "", err
	}
	htmlAddFooter(&html)
	return template.HTML(html.String()), nil
}

// AuditFullHTML returns the audit information from the bid used in the advert
// that resulted in the request to this page.
func (m *PageModel) AuditFullHTML() (template.HTML, error) {
	t, err := m.getTransaction()
	if err != nil {
		return "", err
	}
	if t == nil {
		return template.HTML("<p>Advert not source of request.</p>"), nil
	}

	w, err := m.Winner()
	if err != nil {
		return "", err
	}

	var html bytes.Buffer
	htmlAddHeader(&html)
	err = appendOWIDAndChildren(&html, t, w, 0)
	if err != nil {
		return template.HTML("<p>" + err.Error() + "</p>"), nil
	}
	htmlAddFooter(&html)
	return template.HTML(html.String()), nil
}

func htmlAddHeader(html *bytes.Buffer) {
	html.WriteString("<table class=\"table\">\r\n")
	html.WriteString("<thead>\r\n<tr>\r\n")
	html.WriteString("<th>Organization</th>\r\n")
	html.WriteString("<th>Audit Result</th>\r\n")
	html.WriteString("<th>\r\n</th>\r\n")
	html.WriteString("</tr>\r\n</thead>\r\n<tbody>\r\n")
}

func htmlAddFooter(html *bytes.Buffer) {
	html.WriteString("</tbody>\r\n</table>\r\n")
}

func (m *PageModel) getTransaction() (*owid.OWID, error) {

	if m.offer == nil {

		// Parse the form data.
		err := m.request.ParseForm()
		if err != nil {
			return nil, err
		}

		// If the bid data does not exist return warning.
		if m.request.Form.Get("transaction") == "" {
			return nil, nil
		}

		// Get the transaction from the form data.
		m.offer, err = owid.TreeFromBase64(m.request.Form.Get("transaction"))
		if err != nil {
			return nil, err
		}
	}

	return m.offer, nil
}

func appendParents(html *bytes.Buffer, w *owid.OWID) error {
	var n []*owid.OWID
	p := w
	for p != nil {
		n = append(n, p)
		p = p.GetParent()
	}
	i := len(n) - 1
	for i >= 0 {
		err := appendHTML(html, w, n[i], 0)
		if err != nil {
			return err
		}
		i--
	}
	return nil
}

func appendOWIDAndChildren(
	html *bytes.Buffer,
	o *owid.OWID,
	w *owid.OWID,
	level int) error {
	appendHTML(html, w, o, level)
	if len(o.Children) > 0 {
		for _, c := range o.Children {
			err := appendOWIDAndChildren(html, c, w, level+1)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func appendHTML(
	html *bytes.Buffer,
	w *owid.OWID,
	o *owid.OWID,
	level int) error {

	html.WriteString("<tr>\r\n")
	html.WriteString(fmt.Sprintf(
		"<td style=\"padding-left:%dem;\" class=\"text-left\">\r\n"+
			"<script>new owid().appendName(document.currentScript.parentNode,\"%s\")</script></td>\r\n",
		level,
		o.AsString()))
	html.WriteString(fmt.Sprintf(
		"<td style=\"text-align:center;\">\r\n"+
			"<script>new owid().appendAuditMark(document.currentScript.parentNode,\"%s\",\"%s\");</script>\r\n"+
			"<noscript>JavaScript needed to audit</noscript></td>\r\n",
		o.GetRoot().AsString(),
		o.AsString()))
	if w == o {
		html.WriteString("<td>\r\n<img style=\"width:32px\" src=\"winner.svg\">\r\n</td>")
	} else {
		html.WriteString("<td>\r\n</td>\r\n")
	}
	html.WriteString("</tr>\r\n")
	return nil
}
