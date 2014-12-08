package logtap


import (
  "reflect"
  "time"
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

type Collector interface {
  CollectChan() chan Message
  SetLogger(Logger)
}

type Drain interface {
  Write(Message)
  SetLogger(Logger)
}

type Message struct {
  Time     time.Time
  Priority int
  Content  string
}

type Logtap struct {
  log Logger
  Collectors map[string]Collector
  Drains map[string]Drain
}

// New establishes a new logtap object
// and makes sure it has the some logger
func New(log Logger) *Logtap {
  if log == nil {
    log = DevNullLogger(0)
  }
  return &Logtap{
    log: log,
    Collectors: make(map[string]Collector),
    Drains: make(map[string]Drain),
  }
}

// AddDrain addes a drain to the listeners and sets its logger
func (l *Logtap) AddDrain(tag string, d Drain) {
  d.SetLogger(l.log)
  l.Drains[tag] = d
}

// RemoveDrain drops a drain
func (l *Logtap) RemoveDrain(tag string) {
  delete(l.Drains, tag)
}

// AddCollector adds a collector and begins listening to it
// also adds logging
func (l *Logtap) AddCollector(tag string, c Collector) {
  c.SetLogger(l.log)
  l.Collectors[tag] = c
}

// RemoveCollector remove given collector
func (l *Logtap) RemoveCollector(tag string) {
  delete(l.Collectors, tag)
}

// Start begins listening to all the collectors that are registered.
// it then broadcasts all messages to all the drains that are registered
// this is backgrounded so it can be used by a parent process without getting in the way
func (l *Logtap) Start() {
  go func() {
    for {
      cases := make([]reflect.SelectCase, len(l.Collectors))
      for _, col := range l.Collectors {
        cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(col.CollectChan())})
      }
      // when you append to a nil value [] the first element is the nil value object
      // so its best to remove it
      cases = cases[1:]
      _, value, ok := reflect.Select(cases)
      // ok will be true if the channel has not been closed.
      if ok {
        l.log.Info("[start][collect] %v", value.Interface().(Message))
        l.writeMessage(value.Interface().(Message))
      }
      
    }
    
  }()
}

// writeMessage broadcasts to all drains in seperate go routines
func (l *Logtap) writeMessage(msg Message) {
  for _, drain := range l.Drains {
    go func (d Drain) {
      d.Write(msg)
    }(drain)
  }
}

