package main

import "sync"

type match struct {
	minClients int
	maxClients int
	clients    map[*client]bool
	locker     sync.Locker
	running    bool
	hub        *hub
}

func newMatch(c *client) *match {
	result := &match{
		minClients: 2,
		maxClients: 2,
		clients:    make(map[*client]bool),
		locker:     &sync.Mutex{},
		running:    false,
	}

	result.clients[c] = true
	c.currentMatch = result

	return result
}

// adds client to match if still have space
// if maxClients already reached  returns false and will not add the client
// if maxClient is NOT reached and client was been added returns true
func (m *match) addClient(client *client) bool {
	m.locker.Lock()
	defer func() {
		m.locker.Unlock()
	}()

	if m.maxClients == len(m.clients) {
		m.running = true
		return false
	}

	m.clients[client] = true

	client.currentMatch = m
	return true
}

func (m *match) clientMessage(msg *message) {
	m.locker.Lock()
	defer func() {
		m.locker.Unlock()
	}()

	if !m.running {
		return
	}

	for c := range m.clients {
		if c != msg.sender {
			go func() {
				c.send <- msg
			}()
		}
	}
}

// remove a client from match
// if match is incomplete after remove returns true
// otherwise returns false
func (m *match) removeClient(client *client) bool {
	m.locker.Lock()
	defer func() {
		m.locker.Unlock()
	}()

	delete(m.clients, client)
	client.currentMatch = nil
	if m.minClients > len(m.clients) {
		m.running = false
		return true
	}

	return false
}
