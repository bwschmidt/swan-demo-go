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
	"errors"
	"math/rand"
	"net/http"
	"owid"
	"time"

	"github.com/google/uuid"
)

// Pair contains OWID nad OWID host returned to consumer
type Pair struct {
	Host string `json:"host"`
	Owid string `json:"owid"`
}

// IDs from the request
type IDs struct {
	CBID  *Pair `json:"cbid"`  // Common Browser ID
	SID   *Pair `json:"sid"`   // Signed-in ID
	OID   *Pair `json:"oid"`   // Offer ID
	Allow *Pair `json:"allow"` // Consumer preferences
	TID   *Pair `json:"tid"`   // Transaction ID
}

// Bid payload
type Bid struct {
	Bidder      string `json:"bidder"`      // Bidder ID
	CreativeURL string `json:"creativeURL"` // link to ad
	ClickURL    string `json:"clickURL"`    // link to markter
}

// BidResponse to consumer
type BidResponse struct {
	IDs IDs  `json:"ids"` // ids of companies involved in bid process
	Bid *Bid `json:"bid"` // response
}

func handlerSSP(c *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			returnServerError(c, w, err)
			return
		}

		cbid := r.FormValue("cbid")
		if cbid == "" {
			returnAPIError(
				c,
				w,
				errors.New("cbid param not set"),
				http.StatusBadRequest)
		}
		sid := r.FormValue("sid")
		if sid == "" {
			returnAPIError(
				c,
				w,
				errors.New("sid param not set"),
				http.StatusBadRequest)
		}
		oid := r.FormValue("oid")
		if oid == "" {
			returnAPIError(
				c,
				w,
				errors.New("oid param not set"),
				http.StatusBadRequest)
		}
		allow := r.FormValue("allow")
		if allow == "" {
			returnAPIError(
				c,
				w,
				errors.New("allow param not set"),
				http.StatusBadRequest)
		}

		bid, err := getAd(c, cbid, sid, oid, allow)
		if err != nil {
			returnServerError(c, w, err)
			return
		}
		if err := json.NewEncoder(w).Encode(bid); err != nil {
			returnServerError(c, w, err)
		}

	}
}

func getAd(
	c *Configuration,
	cbid string,
	sid string,
	oid string,
	allow string) (*BidResponse, error) {

	cb, err := getPairFromOwid(cbid)
	if err != nil {
		return nil, err
	}
	s, err := getPairFromOwid(sid)
	if err != nil {
		return nil, err
	}
	o, err := getPairFromOwid(oid)
	if err != nil {
		return nil, err
	}
	a, err := getPairFromOwid(allow)
	if err != nil {
		return nil, err
	}

	ids := IDs{
		cb, s, o, a,
		&Pair{"", uuid.New().String()}}

	keys := []string{}
	for k := range c.Mars {
		keys = append(keys, k)
	}
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(len(keys))
	bidder := keys[randomNum]
	createive := c.Mars[bidder]

	bid := Bid{
		bidder,
		"//" + bidder + createive,
		"//" + bidder + "/mar/"}

	br := BidResponse{ids, &bid}

	return &br, nil
}

func getPairFromOwid(o string) (*Pair, error) {
	var p Pair

	ow, err := owid.DecodeFromBase64(o)
	if err != nil {
		return nil, err
	}

	p.Owid = o
	p.Host = ow.Domain

	return &p, nil
}
