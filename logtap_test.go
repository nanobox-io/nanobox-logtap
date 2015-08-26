package logtap_test

import (
	"github.com/boltdb/bolt"
	"github.com/jcelliott/lumber"
	"github.com/pagodabox/nanobox-logtap/api"
	"github.com/pagodabox/nanobox-logtap/archive"
	"github.com/pagodabox/nanobox-logtap/collector"
	"github.com/pagodabox/nanobox-logtap/drain"
	"os"
)

var log = lumber.NewConsoleLogger(lumber.TRACE)

func TestBasic() {
	logtap.New(log)
	defer logTap.Close()

	console := drain.AdaptWriter(os.Stdout)
	logTap.AddDrain("testing", console)
}

func Bolt() {
	logTap := logtap.New(log)
	defer logTap.Close()

	db, err := bolt.Open("./test.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer func() {
		db.Close()
		os.Remove("./test.db")
	}()

	boltArchive := archive.BoltArchive{
		db:            db,
		maxBucketSize: 10, // store only 10 lines, it is a test.
	}

	logTap.AddDrain("historical", boltArchive)

}
