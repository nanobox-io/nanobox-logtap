// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package logtap_test

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jcelliott/lumber"
	"github.com/pagodabox/golang-hatchet"
	"github.com/pagodabox/nanobox-logtap"
	"github.com/pagodabox/nanobox-logtap/archive"
	"github.com/pagodabox/nanobox-logtap/collector"
	"github.com/pagodabox/nanobox-logtap/drain"
	"net"
	"net/http"
	"os"
	"sync"
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

func TestUDPCollector(test *testing.T) {
	logTap := logtap.New(log)
	defer logTap.Close()
	success := false

	testDrain := func(l hatchet.Logger, msg logtap.Message) {
		success = true
	}

	logTap.AddDrain("drain", testDrain)

	udpCollector, err := collector.SyslogUDPStart("app", "127.0.0.1:1234", logTap)
	assert(test, err == nil, "%v", err)
	defer udpCollector.Close()

	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	assert(test, err == nil, "%v", err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	assert(test, err == nil, "%v", err)

	client, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	assert(test, err == nil, "%v", err)
	defer client.Close()

	_, err = client.Write([]byte("<34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8"))
	assert(test, err == nil, "%v", err)

	time.Sleep(time.Millisecond * 10)
	assert(test, success, "the message was not received")
}

func TestHTTPCollector(test *testing.T) {
	logTap := logtap.New(log)
	defer logTap.Close()
	success := false

	testDrain := func(l hatchet.Logger, msg logtap.Message) {
		success = true
	}

	logTap.AddDrain("drain", testDrain)

	go collector.StartHttpCollector("app", "127.0.0.1:1234", logTap)

	body := bytes.NewReader([]byte("this is a test"))
	res, err := http.Post("http://127.0.0.1:1234/upload", "text/plain", body)
	assert(test, res.StatusCode == 200, "bad response %v", res)
	assert(test, err == nil, "%v", err)
	time.Sleep(time.Millisecond * 10)
	assert(test, success, "the message was not received")
}

func BenchmarkLogvacOne(b *testing.B) {
	benchmarkTest(b, 1)
}

func BenchmarkLogvacTwo(b *testing.B) {
	benchmarkTest(b, 5)
}

func BenchmarkLogvacTen(b *testing.B) {
	benchmarkTest(b, 10)
}

func BenchmarkLogvacOneHundred(b *testing.B) {
	benchmarkTest(b, 100)
}

func benchmarkTest(b *testing.B, listenerCount int) {
	logTap := logtap.New(log)
	defer logTap.Close()

	group := sync.WaitGroup{}
	testDrain := func(l hatchet.Logger, msg logtap.Message) {
		group.Done()
	}

	for i := 0; i < listenerCount; i++ {
		logTap.AddDrain(fmt.Sprintf("%v", i), testDrain)
	}

	group.Add(b.N * listenerCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logTap.Publish("app", lumber.DEBUG, "testing")
	}
	group.Wait()
}

func assert(test *testing.T, check bool, fmt string, args ...interface{}) {
	if !check {
		test.Logf(fmt, args...)
		test.FailNow()
	}
}
