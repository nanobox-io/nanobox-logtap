package logtap

// Overview

import (
	"github.com/pagodabox/golang-hatchet"
	"time"
)

const (
	FATAL = iota
	ERROR
	WARN
	INFO
	DEBUG
)

type (
	Archive interface {
		Slice(name string, offset, limit uint64, level int) ([]Message, uint64, error)
	}

	Drain func(hatchet.Logger, Message)

	Message struct {
		Type     string    `json:"type"`
		Time     time.Time `json:"time"`
		Priority int       `json:"priority"`
		Content  string    `json:"content"`
	}

	Logtap struct {
		log    hatchet.Logger
		drains map[string]drainChannels
	}

	drainChannels struct {
		send chan Message
		done chan bool
	}
)

// Establishes a new logtap object
// and makes sure it has a logger
func New(log hatchet.Logger) *Logtap {
	if log == nil {
		log = hatchet.DevNullLogger{}
	}
	return &Logtap{
		log:    log,
		drains: make(map[string]drainChannels),
	}
}

// Close logtap and remove all drains
func (l *Logtap) Close() {
	for tag := range l.drains {
		l.RemoveDrain(tag)
	}
}

// AddDrain addes a drain to the listeners and sets its logger
func (l *Logtap) AddDrain(tag string, drain Drain) {
	channels := drainChannels{
		done: make(chan bool),
		send: make(chan Message),
	}

	go func() {
		for {
			select {
			case <-channels.done:
				return
			case msg := <-channels.send:
				drain(l.log, msg)
			}
		}
	}()

	l.drains[tag] = channels
}

// RemoveDrain drops a drain
func (l *Logtap) RemoveDrain(tag string) {
	drain, ok := l.drains[tag]
	if ok {
		drain.done <- true
		close(drain.done)
		delete(l.drains, tag)
	}
}

func (l *Logtap) Publish(kind string, priority int, content string) {
	m := Message{
		Type:     kind,
		Time:     time.Now(),
		Priority: priority,
		Content:  content,
	}
	l.WriteMessage(m)
}

// WriteMessage broadcasts to all drains in seperate go routines
// should this wait for the message to be processed by all drains?
func (l *Logtap) WriteMessage(msg Message) {
	for _, drain := range l.drains {
		go func() {
			select {
			case <-drain.done:
				close(drain.send)
				drain.done <- true
			case drain.send <- msg:
			}
		}()
	}
}
