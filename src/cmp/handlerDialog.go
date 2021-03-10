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
	"io/ioutil"
	"net/http"
	"net/url"
	"swift"

	"github.com/google/uuid"
)

type dialogModel struct {
	url.Values         // Key value pairs
	redirectURL string // The URL to redirect the request to
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
	if m.redirectURL != "" {
		http.Redirect(w, r, m.redirectURL, 303)
	} else {
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

	// If this is a close request then don't update the values and just return
	// to the return URL.
	if r.Form.Get("close") != "" {
		m.redirectURL = m.Get("returnUrl")
		return nil
	}

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

	// The user has request that the data be updated in the SWAN network.
	// Set the redirection URL for the operation to store the data. The web
	// browser will then be redirected to that URL, the data saved and the
	// return URL for the publisher returned to.
	m.redirectURL, err = getRedirectUpdateURL(d, &m.Values)

	return err
}

func getRedirectUpdateURL(d *common.Domain, m *url.Values) (string, error) {

	// Build the URL to request the redirect URL for the storage operation.
	var u url.URL
	u.Scheme = d.Config.Scheme
	u.Host = d.SWANAccessNode
	u.Path = "/swan/api/v1/update"
	u.RawQuery = m.Encode()
	q := u.Query()

	// Add the access key for the SWAN network.
	q.Set("accessKey", d.Config.AccessKey)

	// Add the query parameters back with the access key.
	u.RawQuery = q.Encode()

	// Call SWAN to get the URL to redirect the web browser to.
	r, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf(string(b))
	}
	return string(b), nil
}

func decryptAndDecode(d *common.Domain, v string) (*swift.Results, error) {
	var r swift.Results
	var u url.URL
	u.Scheme = d.Config.Scheme
	u.Host = d.SWANAccessNode
	u.Path = "/swan/api/v1/operation-as-json"
	q := u.Query()
	q.Set("accessKey", d.Config.AccessKey)
	q.Set("data", v)
	u.RawQuery = q.Encode()
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(b))
	}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
