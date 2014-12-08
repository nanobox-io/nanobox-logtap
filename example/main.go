package main

import "github.com/nanobox-core/logtap"
import "github.com/jcelliott/lumber"
import "time"

func main() {
  log := lumber.NewConsoleLogger(lumber.INFO)
  log.Prefix("[logtap]")
  ltap := logtap.New(log)
  ltap.Start()

  sysc := logtap.NewSyslogCollector(514)
  l.AddCollector("syslog", sysc)
  sysc.Start()

  conc := logtap.NewConsoleDrain()
  l.AddDrain("concole", conc)

  hist := logtap.NewHistoricalDrain(8080, "./bolt.db", 1000)
  hist.Start()
  l.AddDrain("history", hist)

  // pub := logtap.newPublishDrain(publisher)
  // l.AddDrain("mist", pub)

  time.Sleep(1000*time.Second)
}