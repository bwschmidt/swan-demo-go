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
	"time"
)

var bidJSON = `{
	"ids":[
		"51degrees"
	],
	"bids":[
        {
            "bidder": "cool-cars",
            "creativeUrl": "http://cool-cars.uk:5000/campaign/images/190811762.jpeg",
            "clickUrl": "http://cool-cars.uk:5000/mar"
        },
        {
            "bidder": "cool-bkes",
            "creativeUrl": "http://cool-bikes.uk:5000/campaign/images/234657570.jpeg",
            "clickUrl": "http://cool-bikes.uk:5000/mar"
        },
        {
            "bidder": "cool-creams",
            "creativeUrl": "http://cool-creams.uk:5000/campaign/images/221406343.jpeg",
            "clickUrl": "http://cool-creams.uk:5000/mar"
        }
    ]
}
`

// Bids from the DSP
type Bids struct {
	IDs  []string `json:"ids"`  // ids of companies involved in bid process
	Bids []Bid    `json:"bids"` // response
}

// Bid payload
type Bid struct {
	Bidder      string `json:"bidder"`      // Bidder ID
	CreativeURL string `json:"creativeURL"` // link to ad
	ClickURL    string `json:"clickURL"`    // link to markter
}

// BidResponse to consumer
type BidResponse struct {
	IDs []string `json:"ids"` // ids of companies involved in bid process
	Bid *Bid     `json:"bid"` // response
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

		bid, err := getAd(cbid, sid, oid, allow)
		if err != nil {
			returnServerError(c, w, err)
			return
		}
		if err := json.NewEncoder(w).Encode(bid); err != nil {
			returnServerError(c, w, err)
		}

	}
}

func getAd(cbid string, sid string, oid string, allow string) (*BidResponse, error) {
	var b Bids

	err := json.Unmarshal([]byte(bidJSON), &b)
	if err != nil {
		return nil, err
	}

	count := 0
	for range b.Bids {
		count++
	}

	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(count)

	bid := &b.Bids[randomNum]

	br := BidResponse{append(b.IDs, bid.Bidder), bid}

	return &br, nil
}
