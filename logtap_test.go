// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package logtap_test

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jcelliott/lumber"
	"github.com/pagodabox/golang-hatchet"
	"github.com/pagodabox/nanobox-logtap"
	"github.com/pagodabox/nanobox-logtap/archive"
	"github.com/pagodabox/nanobox-logtap/drain"
	"os"
	"testing"
	"time"
)

var log = lumber.NewConsoleLogger(lumber.TRACE)

func TestBasic(test *testing.T) {
	logTap := logtap.New(log)
	defer logTap.Close()
	called := false

	testDrain := func(l hatchet.Logger, msg logtap.Message) {
		called = true
	}

	console := drain.AdaptWriter(os.Stdout)
	logTap.AddDrain("testing", console)
	logTap.AddDrain("fake", testDrain)
	logTap.Publish("what is this?", lumber.DEBUG, "you should see me!")
	assert(test, called, "the drain was not called")
}

func TestBolt(test *testing.T) {
	logTap := logtap.New(log)
	defer logTap.Close()

	db, err := bolt.Open("./test.db", 0600, nil)
	assert(test, err == nil, "failed to create boltDB %v", err)
	defer func() {
		db.Close()
		os.Remove("./test.db")
	}()

	boltArchive := &archive.BoltArchive{
		DB:            db,
		MaxBucketSize: 10, // store only 10 chunks, this is a test.
	}

	logTap.AddDrain("historical", boltArchive.Write)
	logTap.Publish("app", lumber.DEBUG, "you should see me!")

	// let the other processes finish running
	time.Sleep(100 * time.Millisecond)

	slices, _, err := boltArchive.Slice("app", 0, 100, lumber.DEBUG)
	assert(test, err == nil, "Slice errored %v", err)
	assert(test, len(slices) == 1, "wrong number of slices %v", slices)

	for i := 0; i < 100; i++ {
		logTap.Publish("app", lumber.DEBUG, fmt.Sprintf("log line:%v", i))
	}

	// let the other processes finish running
	time.Sleep(100 * time.Millisecond)

	slices, _, err = boltArchive.Slice("app", 0, 100, lumber.DEBUG)
	assert(test, err == nil, "Slice errored %v", err)
	assert(test, len(slices) == 10, "wrong number of slices %v", len(slices))

}

func assert(test *testing.T, check bool, fmt string, args ...interface{}) {
	if !check {
		test.Logf(fmt, args...)
		test.FailNow()
	}
}
