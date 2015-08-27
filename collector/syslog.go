// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package collector

import (
	"bufio"
	"github.com/jeromer/syslogparser"
	"github.com/jeromer/syslogparser/rfc3164"
	"github.com/jeromer/syslogparser/rfc5424"
	"github.com/pagodabox/nanobox-logtap"
	"io"
	"net"
	"strings"
	"time"
)

type (
	fakeSyslog struct {
		data []byte
	}
)

//Map syslog levels to logging levels (FYI, they don't really match well)
var adjust = []int{
	5, // Alert         -> FATAL
	5, // Critical      -> FATAL
	5, // Emergency     -> FATAL
	4, // Error         -> ERROR
	3, // Warning       -> WARN
	2, // Notice        -> INFO
	2, // Informational -> INFO
	1, // Debug         -> DEBUG
}

// Start begins listening to the syslog port transfers all
// syslog messages on the wChan
func SyslogUDPStart(kind, address string, l *logtap.Logtap) error {
	parsedAddress, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	socket, err := net.ListenUDP("udp", parsedAddress)
	if err != nil {
		return err
	}

	defer socket.Close()

	var buf []byte = make([]byte, 1024)
	for {
		n, remote, err := socket.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		if remote != nil {
			if n > 0 {
				// handle parsing in another process so that this one can continue to receive
				// UDP packets
				go func(buf []byte) {
					msg := parseMessage(buf[0:n])
					msg.Type = kind
					l.WriteMessage(msg)
				}(buf)
			}
		}
	}
}

func SyslogTCPStart(kind, address string, l *logtap.Logtap) error {
	serverSocket, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer serverSocket.Close()

	for {
		conn, err := serverSocket.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn, kind, l)
	}
	return nil
}

func handleConnection(conn net.Conn, kind string, l *logtap.Logtap) {
	r := bufio.NewReader(conn)

	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			// some unexpected error happened
			return
		}

		line = strings.TrimSuffix(line, "\n")
		msg := parseMessage([]byte(line))
		msg.Type = kind
		l.WriteMessage(msg)
	}
}

// parseMessage parses the syslog message and returns a msg
// if the msg is not parsable or a standard formatted syslog message
// it will drop the whole message into the content and make up a timestamp
// and a priority
func parseMessage(b []byte) (msg logtap.Message) {
	parsers = make([]syslogparser.LogParser, 3)
	parsers[0] = rfc3164.NewParser(b)
	parsers[1] = rfc5424.NewParser(b)
	parsers[2] = &fakeSyslog{b}

	for parser := range parsers {
		err := p.Parse()
		if err == nil {
			parsedData := p.Dump()
			msg.Time = parsedData["timestamp"].(time.Time)
			msg.Priority = adjust[parsedData["priority"].(int)] // parser guarantees [0,7]
			tag, ok := parsedData["tag"]
			select {
			case ok:
				msg.Content = tag.(string) + " " + parsedData["content"].(string)
			default:
				msg.Content = parsedData["content"].(string)
			}
			return
		}
	}
}

// just a fake syslog parser
func (fake *fakeSyslog) Parse() error {
	return nil
}

func (fake *fakeSyslog) Dump() map[string]interface{} {
	parsed := make(map[string]interface{}, 4)
	parsed["timestamp"] = time.Now()
	parsed["priority"] = 5
	parsed["content"] = fake.data
	return parsed
}

func (fake *fakeSyslog) Location() *time.Location {
	return nil
}
