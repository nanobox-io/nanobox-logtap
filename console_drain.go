package logtap



import "fmt"


type ConsoleDrain int


func NewConsoleDrain() ConsoleDrain {
  return ConsoleDrain(0)
}

func (c ConsoleDrain) Write(msg Message) {
  fmt.Println(msg.Time, msg.Priority, msg.Content)
}