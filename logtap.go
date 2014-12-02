package logtap


import (
  "net/http"
  "io/ioutil"
  "strconv"

)

type Logger interface {
  Fatal(string, ...interface{})
  Error(string, ...interface{})
  Warn(string, ...interface{})
  Info(string, ...interface{})
  Debug(string, ...interface{})
  Trace(string, ...interface{})
}

type DevNullLogger int8
func (d DevNullLogger) Fatal(thing string,v ...interface{}) {}
func (d DevNullLogger) Error(thing string,v ...interface{}) {}
func (d DevNullLogger) Warn(thing string,v ...interface{}) {}
func (d DevNullLogger) Info(thing string,v ...interface{}) {}
func (d DevNullLogger) Debug(thing string,v ...interface{}) {}
func (d DevNullLogger) Trace(thing string,v ...interface{}) {}

type Logtap struct {
  log Logger

  Port   int
  Drains []Drain
}


func New(port int, log Logger) *Logtap {
  if log == nil {
    log = DevNullLogger(0)
  }
  return &Logtap{
    log: log,
    Port: port,
  }
}

func (l *Logtap) AddDrain(d Drain) {
  l.Drains << d
}

func (l *Logtap) RemoveDrain(d Drain) {
  
}

func (l *Logtap) start() {
  
}


