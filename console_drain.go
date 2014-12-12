package logtap

import "fmt"

type ConsoleDrain struct {
  log Logger
}

// NewConcoleDrain creates a new drain and uses a devnull logger
func NewConsoleDrain() *ConsoleDrain {
  return &ConsoleDrain{DevNullLogger(0)}
}

// SetLogger really allows the logtap main struct
// to assign its own logger to the concole drain
func (c *ConsoleDrain) SetLogger(l Logger) {
  c.log = l
}

func (c *ConsoleDrain) Write(msg Message) {
  c.log.Info("[concole][write] message:"+msg.Content)
  fmt.Printf("[%s] <%d> %s", msg.Time, msg.Priority, msg.Content)
}