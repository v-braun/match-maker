package main

import "sync"

type Match struct {
	minClients int
	maxClients int
	clients    map[*Client]bool
	locker     sync.Locker
}

func NewMatch(client *Client) *Match {
	result := &Match{
		minClients: 2,
		maxClients: 2,
		clients:    make(map[*Client]bool),
		locker:     &sync.Mutex{},
	}

	result.clients[client] = true

	return result
}

// adds client to match if still have space
// if maxClients already reached  returns false and will not add the client
// if maxClient is NOT reached and client was been added returns true
func (m *Match) addClient(client *Client) bool {
	m.locker.Lock()
	defer func() {
		m.locker.Unlock()
	}()

	if m.maxClients == len(m.clients) {
		return false
	}

	m.clients[client] = true

	client.currentMatch = m
	return true
}

func (m *Match) clientMessage(msg *Message) {

}

// remove a client from match
// if match is incomplete after remove returns true
// otherwise returns false
func (m *Match) removeClient(client *Client) bool {
	m.locker.Lock()
	defer func() {
		m.locker.Unlock()
	}()

	delete(m.clients, client)
	client.currentMatch = nil
	if m.minClients > len(m.clients) {
		return true
	}

	return false
}
