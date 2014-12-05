package main

import "github.com/nanobox-core/logtap"
import "github.com/jcelliott/lumber"
import "time"

func main() {
  log := lumber.NewConsoleLogger(lumber.ERROR)

  r := logtap.New(514, log)
  r.Start()
  // d := logtap.NewConsoleDrain()
  d2 := logtap.NewHistoricalDrain(8080, "./bolt.db", 1000)
  d2.Start()
  // r.AddDrain("concole", d)
  r.AddDrain("history", d2)
  time.Sleep(1000*time.Second)
}