package logtap

import "fmt"
import "github.com/pagodabox/golang-hatchet"

type Publisher interface {
	Publish(tags []string, data string)
}

type PublishDrain struct {
	log       hatchet.Logger
	publisher Publisher
}

// NewPublishDrain creates a new publish drain and returns it
func NewPublishDrain(pub Publisher) *PublishDrain {
	return &PublishDrain{
		publisher: pub,
	}
}

// SetLogger really allows the logtap main struct
// to assign its own logger to the publsih drain
// the publsih drain doesnt use the logger but
// it is necessary to have the method to match the interface
// the assumption here is that the publisher will do its own loggin
func (p *PublishDrain) SetLogger(l hatchet.Logger) {
	p.log = l
}

// Write formats the data coming in on the message and drops it on the publish method
// in a format the publisher can use
func (p *PublishDrain) Write(msg Message) {
	p.log.Debug("[LOGTAP][publish][write]:[%s] %s", msg.Time, msg.Content)
	tags := []string{"log", msg.Type}
	severities :=[]string{"emergency","alert","critical","error","warning","notice","informational","debug"}
	tags = append(tags, severities[msg.Priority:]...)
	p.publisher.Publish(tags, fmt.Sprintf("{\"time\":\"%s\",\"log\":\"%s\"}", msg.Time, msg.Content))
}
