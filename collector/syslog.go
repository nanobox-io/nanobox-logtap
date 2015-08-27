// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package collector

import (
	"bufio"
	"github.com/jeromer/syslogparser/rfc3164"
	"github.com/jeromer/syslogparser/rfc5424"
	"github.com/pagodabox/nanobox-logtap"
	"io"
	"net"
	"strings"
	"time"
)

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

		// handle each connection individually (non-blocking)
		go handleConnection(conn, kind, l)
	}
	return nil
}

func handleConnection(conn net.Conn, kind string, l *logtap.Logtap) {
	r := bufio.NewReader(conn)

	//
	for {

		// read messages coming across the tcp channel
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
	p := rfc3164.NewParser(b)
	err := p.Parse()
	if err == nil {
		parsedData := p.Dump()
		// fmt.Printf("%#v\n",parsedData)
		msg.Time = parsedData["timestamp"].(time.Time)
		msg.Priority = adjustInt(parsedData["priority"].(int) % 8)
		msg.Content = parsedData["tag"].(string) + " " + parsedData["content"].(string)
	} else {
		p := rfc5424.NewParser(b)
		err := p.Parse()
		if err == nil {
			parsedData := p.Dump()
			// fmt.Printf("%#v\n",parsedData)
			msg.Time = parsedData["timestamp"].(time.Time)
			msg.Priority = adjustInt(parsedData["priority"].(int) % 8)
			msg.Content = parsedData["tag"].(string) + " " + parsedData["content"].(string)
		} else {
			msg.Time = time.Now()
			msg.Priority = 1
			msg.Content = string(b)
		}
	}
	return
}

// I need to adjust the possible prioritys from rfc3164 and rfc5424
// to the 5 priority options.
func adjustInt(in int) int {
	if in < 3 {
		return 0
	}
	if in < 5 {
		return in - 2
	}
	return in - 3
}
