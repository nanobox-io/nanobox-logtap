package logtap

import "fmt"

type Publisher interface{
  Publish(tags []string, data string)
}


type PublishDrain struct {
  publisher Publisher
}


func NewPublishDrain(pub Publisher) PublishDrain {
  return PublishDrain{
    publisher: pub,
  }
}

func (p *PublishDrain) Write(msg Message) {
  p.publisher.Publish([]string{"log"}, fmt.Sprintf("[%s] %s", msg.Time, msg.Content))
}