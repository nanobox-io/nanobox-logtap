package api

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/pagodabox/golang-hatchet"
	"github.com/pagodabox/nanobox-logtap"
	"net/http"
	"strconv"
)

func GenerateArchiveEndpoint(archive logtap.Archive) http.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		code := 200
		var body []byte

		// where do these come from?
		name := r.FormValue("kind")
		offset := r.FormValue("offset", 0)
		limit := r.FormValue("limit", 100)
		level := r.FormValue("level", 6)

		slices, nextIdx, err := archive.Slice(name, offset, limit, level)
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
