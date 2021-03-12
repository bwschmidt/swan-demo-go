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

package cmp

import (
	"common"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"swift"

	"github.com/google/uuid"
)

type dialogModel struct {
	url.Values      // Key value pairs
	update     bool // True if the update should be performed
}

func (m *dialogModel) Title() string           { return m.Get("title") }
func (m *dialogModel) CBID() string            { return m.Get("cbid") }
func (m *dialogModel) Email() string           { return m.Get("email") }
func (m *dialogModel) Allow() string           { return m.Get("allow") }
func (m *dialogModel) BackgroundColor() string { return m.Get("backgroundColor") }
func (m *dialogModel) PublisherHost() string {
	u, _ := url.Parse(m.Get("returnUrl"))
	if u != nil {
		return u.Host
	}
	return ""
}

func handlerDialog(d *common.Domain, w http.ResponseWriter, r *http.Request) {
	var m dialogModel

	// Parse the form variables.
	err := r.ParseForm()
	if err != nil {
		common.ReturnServerError(d.Config, w, err)
		return
	}

	// Set the storage operation data form the URL in the dialog model.
	err = dialogGetModel(d, r, &m)
	if err != nil {
		common.ReturnServerError(d.Config, w, err)
		return
	}

	// If this is a close request then don't update the values and just return
	// to the return URL.
	if r.Form.Get("close") != "" {
		http.Redirect(w, r, m.Get("returnUrl"), 303)
		return
	}

	// If the method is POST then update the model with the data from the form.
	if r.Method == "POST" {
		err = dialogUpdateModel(d, r, &m)
		if err != nil {
			common.ReturnServerError(d.Config, w, err)
			return
		}
	}

	// If the redirect URL has been set then redirect, otherwise display the
	// HTML template.
	if m.update == true {

		// The user has request that the data be updated in the SWAN network.
		// Set the redirection URL for the operation to store the data. The web
		// browser will then be redirected to that URL, the data saved and the
		// return URL for the publisher returned to.
		u, err := getRedirectUpdateURL(d, r, m.Values)
		if err != nil {
			common.ReturnProxyError(d.Config, w, err)
		}
		http.Redirect(w, r, u, 303)

	} else {

		// The dialog needs to be displayed. Use the cmp.html template for the
		// user interface.
		err := d.LookupHTML("cmp.html").Execute(w, &m)
		if err != nil {
			common.ReturnServerError(d.Config, w, err)
			return
		}
	}
}

func dialogGetModel(d *common.Domain,
	r *http.Request,
	m *dialogModel) error {
	m.Values = make(url.Values)

	// Get the SWAN data from the request path.
	s := common.GetSWANDataFromRequest(r)
	if s == "" {
		return fmt.Errorf(
			"Path '%s' does not contain SWAN data",
			r.URL.Path)
	}

	// Call the SWAN access node for the CMP to turn the data provided in the
	// URL into usable data for the dialog.
	op, err := decryptAndDecode(d, s)
	if err != nil {
		return err
	}

	// Build the form parameters from the data received from SWAN.
	m.Set("title", op.HTML.Title)
	m.Set("backgroundColor", op.HTML.BackgroundColor)
	m.Set("messageColor", op.HTML.MessageColor)
	m.Set("progressColor", op.HTML.ProgressColor)
	m.Set("message", op.HTML.Message)
	m.Set("returnUrl", op.State) // State is the return URL for the dialog.
	if op.Get("cbid") != nil && op.Get("cbid").Value != "" {
		m.Set("cbid", op.Get("cbid").Value)
	} else {
		m.Set("cbid", uuid.New().String())
	}
	if op.Get("email") != nil {
		m.Set("email", op.Get("email").Value)
	}
	if op.Get("allow") != nil {
		m.Set("allow", op.Get("allow").Value)
	}

	return nil
}

func dialogUpdateModel(
	d *common.Domain,
	r *http.Request,
	m *dialogModel) error {
	var err error

	// Copy the field values from the form.
	m.Values.Set("cbid", r.Form.Get("cbid"))
	m.Values.Set("email", r.Form.Get("email"))
	m.Values.Set("allow", r.Form.Get("allow"))

	// Check to see if the post is as a result of the CBID reset.
	if r.Form.Get("reset-cbid") != "" {

		// Replace the CBID with a new random value.
		m.Set("cbid", uuid.New().String())
		return nil
	}

	// Check to see if the post is as a result for all data.
	if r.Form.Get("reset-all") != "" {

		// Replace the data.
		m.Set("email", "")
		m.Set("allow", "")
		m.Set("cbid", uuid.New().String())
		return nil
	}

	// The data should be updated in the SWAN network.
	m.update = true

	return err
}

func getRedirectUpdateURL(
	d *common.Domain,
	r *http.Request,
	m url.Values) (string, *common.SWANError) {
	b, err := d.CallSWANURL("update", func(q *url.Values) error {
		for k, v := range m {
			if k == "allow" && v[0] == "" {
				q.Add(k, "off")
			} else {
				for _, i := range v {
					q.Add(k, i)
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func decryptAndDecode(d *common.Domain, v string) (
	*swift.Results,
	*common.SWANError) {
	var r swift.Results
	b, e := d.CallSWANURL("operation-as-json", func(q *url.Values) error {
		q.Set("data", v)
		return nil
	})
	if e != nil {
		return nil, e
	}
	err := json.Unmarshal(b, &r)
	if err != nil {
		return nil, &common.SWANError{err, nil}
	}
	return &r, nil
}
