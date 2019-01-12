package main

import "encoding/binary"

type messageType uint32

const wantMatchMaking messageType = 1
const wantLeaveMatch messageType = 3

const matchMessage messageType = 2

//const LeaveMatch = 3

type message struct {
	msgType messageType
	data    []byte
	sender  *client
}

func newMessage(client *client, data []byte) *message {
	t := messageType(binary.BigEndian.Uint32(data))
	rest := data[4:]
	sender := client
	result := &message{
		msgType: t,
		data:    rest,
		sender:  sender,
	}

	return result
}

func (m *message) toData() []byte {
	typeBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(typeBuffer, uint32(m.msgType))

	payload := append(typeBuffer, m.data...)

	return payload
}
