package collector

import (
	"github.com/jcelliott/lumber"
	"github.com/pagodabox/nanobox-logtap"
	"io/ioutil"
	"net/http"
)

// create and return a http handler that can be dropped into an api.
func NewHttpCollector(kind string, l *logtap.Logtap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		l.Publish(kind, lumber.LvlInt(r.Header.Get("X-Log-Level")), string(body))
	}
}

func Start(kind, address string, l *logtap.Logtap) error {
	return http.ListenAndServe(address, NewHttpCollector(kind, l))
}
