package main

import "encoding/binary"

type MessageType uint32

type Message struct {
	msgType MessageType
	data    []byte
	sender  *Client
}

func (m *Message) toData() []byte {
	typeBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(typeBuffer, uint32(m.msgType))

	payload := append(typeBuffer, m.data...)

	return payload
}
