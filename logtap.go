package logtap


import (
  "strconv"
  "net"
  "github.com/jeromer/syslogparser/rfc3164"
  "github.com/jeromer/syslogparser/rfc5424"
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

type Drain interface {
  Write(Message)
}

type Message struct {
  Time     time.Time
  Priority int
  Content  string
}

type Logtap struct {
  log Logger

  Port   int
  Drains map[string]Drain
}


func New(port int, log Logger) *Logtap {
  if log == nil {
    log = DevNullLogger(0)
  }
  return &Logtap{
    log: log,
    Port: port,
    Drains: make(map[string]Drain),
  }
}

func (l *Logtap) AddDrain(tag string, d Drain) {
  l.Drains[tag] = d
}

func (l *Logtap) RemoveDrain(tag string, d Drain) {
  delete(l.Drains, tag)
}

func (l *Logtap) Start() {
  go func () {
    udpAddress, err := net.ResolveUDPAddr("udp4",("0.0.0.0:"+strconv.Itoa(l.Port)))
    if err != nil {
      l.log.Error("error resolving UDP address on ", l.Port)
      l.log.Error(err.Error())
      return
    }

    conn, err := net.ListenUDP("udp", udpAddress)
    if err != nil {
      l.log.Error("error listening on UDP port ", l.Port)
      l.log.Error(err.Error())
      return
    }
    defer conn.Close()

    var buf []byte = make([]byte, 1024)
    for {
      n, address, err := conn.ReadFromUDP(buf)
      if err != nil {
        l.log.Error("error reading data from connection")
        l.log.Error(err.Error())
      }
      if address != nil {
        l.log.Info("got message from "+address.String()+" with n = "+strconv.Itoa(n))
        if n > 0 {
          msg := l.ParseMessage(buf[0:n])
          l.log.Info("msg content: "+msg.Content)
          l.writeMessage(msg)
        }
      }
    }
  }()

}

func (l *Logtap) writeMessage(msg Message) {
  for _, drain := range l.Drains {
    go func () {
      drain.Write(msg)
    }()
  }
}

func (l *Logtap) ParseMessage(b []byte) (msg Message) {
  p := rfc3164.NewParser(b)
  err := p.Parse()
  if err == nil {
    parsedData := p.Dump()
    msg.Time = parsedData["timestamp"].(time.Time)
    msg.Priority = parsedData["priority"].(int)
    msg.Content = parsedData["content"].(string)
  } else {
    p := rfc5424.NewParser(b)
    err := p.Parse()
    if err == nil {
      parsedData := p.Dump()
      msg.Time = parsedData["timestamp"].(time.Time)
      msg.Priority = parsedData["priority"].(int)
      msg.Content = parsedData["content"].(string)
    } else {
      l.log.Error("Unable to parse data: "+string(b))
      msg.Time = time.Now()
      msg.Priority = 1
      msg.Content = string(b)
    }
  }
  return
}
