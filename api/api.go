// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package api

import (
	"encoding/json"
	"github.com/pagodabox/nanobox-logtap"
	"net/http"
)

func GenerateArchiveEndpoint(archive logtap.Archive) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		code := 200
		var body []byte

		// where do these come from?
		// name := req.FormValue("kind")
		// offset := req.FormValue("offset")
		// limit := req.FormValue("limit")
		// level := req.FormValue("level")

		slices, _, err := archive.Slice("app", 0, 100, 10)
		if err != nil {
			code = 500
			body = []byte(err.Error())
		} else {
			body, err = json.Marshal(slices)
			if err != nil {
				code = 500
				body = []byte(err.Error())
			}
		}

		res.WriteHeader(code)
		res.Write(body)
	}
}
