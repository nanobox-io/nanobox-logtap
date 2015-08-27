// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package main

import (
	"github.com/boltdb/bolt"
	"github.com/jcelliott/lumber"
	"github.com/pagodabox/nanobox-logtap"
	"github.com/pagodabox/nanobox-logtap/api"
	"github.com/pagodabox/nanobox-logtap/archive"
	"github.com/pagodabox/nanobox-logtap/collector"
	"github.com/pagodabox/nanobox-logtap/drain"
)

func main() {
	log := lumber.NewConsoleLogger(lumber.INFO)
	log.Prefix("[logtap]")

	logTap := logtap.New(log)

	db, err := bolt.Open("./test.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	boltArchive := archive.BoltArchive{
		db:            db,
		maxBucketSize: 10, // store only 10 lines
	}

	// add in all the drains that we have created
	logTap.AddDrain("historical", boltArchive)
	logTap.AddDrain("mist", mist)

	sysc := logtap.NewSyslogCollector("514")
	ltap.AddCollector("syslog", sysc)
	sysc.Start()

	post := logtap.NewHttpCollector("6361")
	ltap.AddCollector("post", post)
	post.Start()

	conc := logtap.NewConsoleDrain()
	ltap.AddDrain("concole", conc)

	hist := logtap.NewHistoricalDrain("8080", "./bolt.db", 1000)
	hist.Start()
	ltap.AddDrain("history", hist)

	// pub := logtap.newPublishDrain(publisher)
	// l.AddDrain("mist", pub)
	time.Sleep(1000 * time.Second)
}
