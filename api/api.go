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
