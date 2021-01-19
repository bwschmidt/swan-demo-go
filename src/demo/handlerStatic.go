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
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

// handleStatic locates and returns static content if relevant to the HTTP
// request. True is returned if static content was returned, otherwise false.
func handleStatic(
	d *Domain,
	w http.ResponseWriter,
	r *http.Request) (bool, error) {
	var err error
	folder := d.folder
	found := false
	for found == false && strings.Contains(folder, "www") {
		found, err = handleStaticFolder(d, w, r, folder)
		if err != nil {
			return false, err
		}
		if found == false {
			folder = filepath.Dir(folder)
		}
	}
	return found, nil
}

func handleStaticFolder(
	d *Domain,
	w http.ResponseWriter,
	r *http.Request,
	folder string) (bool, error) {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return false, err
	}
	for _, f := range files {
		if f.IsDir() == false &&
			f.Name() == filepath.Base(r.URL.Path) {
			return handlerFile(w, r, filepath.Join(folder, f.Name()))
		}
	}
	return false, nil
}

func handlerFile(
	w http.ResponseWriter,
	r *http.Request,
	f string) (bool, error) {
	switch filepath.Ext(f) {
	case ".ico":
		http.ServeFile(w, r, f)
		return true, nil
	case ".jpeg":
		http.ServeFile(w, r, f)
		return true, nil
	case ".jpg":
		http.ServeFile(w, r, f)
		return true, nil
	case ".png":
		http.ServeFile(w, r, f)
		return true, nil
	case ".css":
		http.ServeFile(w, r, f)
		return true, nil
	case ".js":
		http.ServeFile(w, r, f)
		return true, nil
	case ".svg":
		http.ServeFile(w, r, f)
		return true, nil
	case ".map":
		http.ServeFile(w, r, f)
		return true, nil
	}
	return false, nil
}
