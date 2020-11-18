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
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"owid"
	"swan"
	"swift"
)

// AddHandlers and outputs configuration information.
func AddHandlers(settingsFile string) {

	// Get the demo configuration.
	dc := newConfig(settingsFile)

	// Get the example simple access control implementations.
	swi := swift.NewAccessSimple(dc.AccessKeys)
	oa := owid.NewAccessSimple(dc.AccessKeys)
	swa := swan.NewAccessSimple(dc.AccessKeys)

	// Add the SWAN handlers, with the publisher handler being used for any
	// malformed storage requests.
	swan.AddHandlers(settingsFile, swa, swi, oa, handlerPublisher(&dc))

	// TODO Add a handler for the marketers end point.
	// http.HandleFunc("/mar", handlerPublisher(&demoConfig))

	// Start the web server on the port provided.
	log.Printf("Demo scheme: %s\n", dc.Scheme)
	log.Printf("SWAN access node domain: %s\n", dc.SwanDomain)
	log.Println("Pub. URLs:")
	for _, s := range dc.Pubs {
		log.Println("  ", s)
	}
	log.Println("Mar. URLs:")
	for _, s := range dc.Mars {
		log.Println("  ", s)
	}
}

func newResponseError(url string, resp *http.Response) error {
	in, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return fmt.Errorf("API call '%s' returned '%d' and '%s'",
		url, resp.StatusCode, in)
}

func returnServerError(c *Configuration, w http.ResponseWriter, err error) {
	w.Header().Set("Cache-Control", "no-cache")
	if c.Debug {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		http.Error(w, "", http.StatusInternalServerError)
	}
	if c.Debug {
		println(err.Error())
	}
}

func getBackgroundColor(d string) string {
	h := fnv.New32a()
	h.Write([]byte(d))
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(h.Sum32()))
	return fmt.Sprintf("#%x%x%x", b[0], b[1], b[2])
}
