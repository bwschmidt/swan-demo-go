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
	"io/ioutil"
	"math/rand"
	"net/http"
	"owid"
	"swan"
)

// handleTransaction is responsible for a real time transaction for advertising.
// The body of the request must contain a JSON array of Processor IDs which
// contain the signature of the last entry in the list.
func handleTransaction(
	d *Domain,
	w http.ResponseWriter,
	r *http.Request) (bool, error) {
	if r.URL.Path == "/demo/api/v1/bid" && r.Method == "POST" && d.owid != nil {

		// Unpack the body of the request to form the bid data structure.
		o, err := getOffer(r)
		if err != nil {
			return true, err
		}

		// If this domain is a bad actor then change the publisher's domain to
		// one that would generate more money from advertising.
		if d.Bad {
			err = changePubDomain(o, "high-value-pub.com")
			if err != nil {
				return true, err
			}
		}

		// Handle the bid and return if the URL was found.
		t, err := handleBid(d, o)
		if err != nil {
			return true, err
		}

		// The caller already knows about the rest of the tree. Only return this
		// Processor OWID and the children.
		b, err := t.TreeAsByteArray()
		if err != nil {
			return true, err
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(b)

		return true, err
	}
	return false, nil
}

func changePubDomain(r *owid.OWID, newPubDomain string) error {
	fake, err := swan.OfferFromOWID(r)
	if err != nil {
		return err
	}
	fake.PubDomain = newPubDomain
	r.Payload, err = fake.AsByteArray()
	if err != nil {
		return err
	}
	return nil
}

func handleBid(d *Domain, o *owid.OWID) (*owid.OWID, error) {

	// The single leaf is the parent Processor OWID. If there isn't a single
	// leaf then too much information has been sent from the caller.
	parent, err := o.GetLeaf()
	if err != nil {
		return nil, err
	}

	// Create an OWID for this processor.
	t := d.owid.CreateOWID(nil)
	if t == nil {
		return nil, fmt.Errorf("Could not create new OWID")
	}

	// If this domain has adverts then choose one at random. Get a random
	// byte array to use as the payload from the Processor OWID.
	if len(d.Adverts) > 0 {
		w := d.Adverts[rand.Intn(len(d.Adverts))]
		var b swan.Bid
		b.AdvertiserURL = w.AdvertiserURL
		b.MediaURL = w.MediaURL
		t.Payload, err = b.AsByteArray()
		if err != nil {
			return nil, err
		}
	}

	// Add this Processor OWID to the children of the parent.
	_, err = parent.AddChild(t)
	if err != nil {
		return nil, err
	}

	// Sign the Processor OWID now that it's part of the tree. This can be used
	// by down stream suppliers to verify that this processor was involved in
	// the transaction.
	err = d.owid.Sign(t)
	if err != nil {
		return nil, err
	}

	// Call all the suppliers adding them to this Processor OWID's child
	// transactions.
	c := make([]*owid.OWID, len(d.Suppliers))
	for i, s := range d.Suppliers {
		c[i], err = sendToSupplier(d.config.Scheme+"://"+s, o)
		if err != nil {
			return nil, err
		}
	}

	// Merge the results from the suppliers.
	t.AddChildren(c)

	// If there are children then pick one at random for the payload of
	// this processor. Used to determine the winner when the transaction
	// is complete. This also demonstrates how the payload can be changed
	// after the response has been received.
	if len(t.Children) > 0 {
		w := rand.Intn(len(t.Children))
		writeUint32(t, uint32(w))
		err = d.owid.Sign(t)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func getOffer(r *http.Request) (*owid.OWID, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return owid.TreeFromByteArray(b)
}

func sendToSupplier(u string, o *owid.OWID) (*owid.OWID, error) {

	// Turn the Offer OWID into a byte array.
	d, err := o.TreeAsByteArray()
	if err != nil {
		return nil, err
	}

	// POST the bid to the supplier.
	up := u + "/demo/api/v1/bid"
	res, err := http.Post(
		up,
		"application/octet-stream",
		bytes.NewBuffer(d))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(u, res)
	}

	// Read the response as a byte array.
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Convert the byte array to a response OWID.
	s, err := owid.TreeFromByteArray(b)
	if err != nil {
		fmt.Println(string(b))
		return nil, fmt.Errorf("Invalid response from '%s'", up)
	}

	return s, nil
}
