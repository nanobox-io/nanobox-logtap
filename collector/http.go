package collector

import (
	"github.com/pagodabox/nanobox-logtap"
	"net/http"
)

// create and return a http handler that can be dropped into an api.
func NewHttpCollector(kind string, l *Logtap) http.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		l.Publish(kind, priorityInt(r.Header.Get("X-Log-Level")), string(body))
	}
}

func Start(kind, address string) error {
	return http.ListenAndServe(address, NewHttpCollector(kind))
}
