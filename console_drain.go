package logtap

import "fmt"

type ConsoleDrain struct {
  log Logger
}

func NewConsoleDrain() *ConsoleDrain {
  return &ConsoleDrain{DevNullLogger(0)}
}

func (c *ConsoleDrain) SetLogger(l Logger) {
  c.log = l
}

func (c *ConsoleDrain) Write(msg Message) {
  c.log.Info("[concole][write] message:"+msg.Content)
  fmt.Printf("[%s] <%d> %s", msg.Time, msg.Priority, msg.Content)
}