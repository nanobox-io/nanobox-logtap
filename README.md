# Logtap

Logtap is an embeddable and configurable log aggregation, storage, and publishing service.

## Memory Usage

## logtap.Drain

A `logtap.Drain` is a simple endpoint that accepts logs that are sent through logtap. Multiple drains can be created and added to logvac. A drain can represent logs that are streamed to stdout, a file, a tcp socket, or anything that can be wrapped to accept `logtap.Message` structs. There are 3 adapters stored at `logtap/drain.Adapt*` that can be used to adapt common interfaces to the logvac.Drain interface.

## logtap.Archive

An Archive is an interface for retreiving a slice of logs from the opaque storage medium. Currently there exists one storage option: BoltDB.

## Example

This example will create a logtap that accepts udp syslog packets and stores them on disk in a file called `./temp.log`.

```go

package main

import (
  "github.com/pagodabox/nanobox-logtap"
  "github.com/pagodabox/nanobox-logtap/drain"
  "github.com/pagodabox/nanobox-logtap/collector"
  "os"
  "os/signal"
)

func main(){
  logTap := logtap.New(nil)
  defer logTap.Close()
  
  file, err := os.Create("./temp.log")
  if err != nil {
    fatal(err)
  }
  defer file.Close()

  logTap.AddDrain("file", drain.AdaptWriter(file))

  udpCollector, err := collector.SyslogUDPStart("app-logs", ":514" ,logTap)
  if err != nil {
    fatal(err)
  }
  defer udpCollector.Close()

  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt, os.Kill)

  // wait for a signal to arrive
  s := <-c
}
```


### Notes: