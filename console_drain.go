package logtap

import "fmt"
import "github.com/nanobox-core/hatchet"

type ConsoleDrain struct {
	log hatchet.Logger
}

// NewConcoleDrain creates a new drain and uses a devnull logger
func NewConsoleDrain() *ConsoleDrain {
	return &ConsoleDrain{}
}

// SetLogger really allows the logtap main struct
// to assign its own logger to the concole drain
func (c *ConsoleDrain) SetLogger(l hatchet.Logger) {
	c.log = l
}

// Write formats the message given and prints it to stdout
func (c *ConsoleDrain) Write(msg Message) {
	c.log.Info("[concole][write] message:" + msg.Content)
	fmt.Printf("[%s] <%d> %s", msg.Time, msg.Priority, msg.Content)
}
