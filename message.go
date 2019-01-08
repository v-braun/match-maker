package main

import "encoding/binary"

type MessageType uint32

const WantMatchMaking MessageType = 1
const WantLeaveMatch MessageType = 3

const MatchMessage MessageType = 2

//const LeaveMatch = 3

type Message struct {
	msgType MessageType
	data    []byte
	sender  *Client
}

func NewMessage(client *Client, data []byte) *Message {
	t := MessageType(binary.BigEndian.Uint32(data))
	rest := data[4:]
	sender := client
	result := &Message{
		msgType: t,
		data:    rest,
		sender:  sender,
	}

	return result
}

func (m *Message) toData() []byte {
	typeBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(typeBuffer, uint32(m.msgType))

	payload := append(typeBuffer, m.data...)

	return payload
}
