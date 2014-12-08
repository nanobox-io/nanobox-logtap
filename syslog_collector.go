package logtap

import (
  "strconv"
  "net"
  "github.com/jeromer/syslogparser/rfc3164"
  "github.com/jeromer/syslogparser/rfc5424"
  "time"
)

type SyslogCollector struct {
  wChan chan Message
  log Logger

  Port   int
}

func NewSyslogCollector(port int) *SyslogCollector {
  return &SyslogCollector{
    log: DevNullLogger(0),
    Port: port,
    wChan: make(chan Message),
  }
}

func (s *SyslogCollector) SetLogger(l Logger) {
  s.log = l
}

func (s *SyslogCollector) CollectChan() chan Message {
  return s.wChan
}

func (s *SyslogCollector) Start() {
  go func () {

    udpAddress, err := net.ResolveUDPAddr("udp4",("0.0.0.0:"+strconv.Itoa(s.Port)))
    if err != nil {
      s.log.Error("error resolving UDP address on ", s.Port)
      s.log.Error(err.Error())
      return
    }

    conn, err := net.ListenUDP("udp", udpAddress)
    if err != nil {
      s.log.Error("error listening on UDP port ", s.Port)
      s.log.Error(err.Error())
      return
    }
    defer conn.Close()


    var buf []byte = make([]byte, 1024)
    for {
      // s.log.Info("[syslog][start] listen")
      n, address, err := conn.ReadFromUDP(buf)
      if err != nil {
        s.log.Error("error reading data from connection")
        s.log.Error(err.Error())
      }
      if address != nil {
        // s.log.Info("[syslog][start] got message from "+address.String()+" with n = "+strconv.Itoa(n))
        if n > 0 {
          msg := s.ParseMessage(buf[0:n])
          s.log.Info("[syslog][start] msg content: "+msg.Content)
          s.wChan <- msg
        }
      }
    }
  }()
}

func (s *SyslogCollector) ParseMessage(b []byte) (msg Message) {
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
      s.log.Error("Unable to parse data: "+string(b))
      msg.Time = time.Now()
      msg.Priority = 1
      msg.Content = string(b)
    }
  }
  return
}
