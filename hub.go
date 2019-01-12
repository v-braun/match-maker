package main

import "sync"

type hub struct {
	wg sync.WaitGroup

	// Registered clients.
	clients map[*client]bool

	// Inbound messages from the clients.
	clientMessage chan *message

	// register new client in the hub
	register chan *client

	// unregister client in the hub
	unregister chan *client

	pendingMatch *match

	close chan interface{}
}

func startHub() *hub {
	h := &hub{
		clientMessage: make(chan *message),
		unregister:    make(chan *client),
		register:      make(chan *client),
		clients:       make(map[*client]bool),
	}

	go h.runClientIO()

	return h
}

func (h *hub) runClientIO() {

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.wg.Add(1)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				match := client.currentMatch
				if match != nil {
					match.removeClient(client)
				}
				h.wg.Done()
			}
		case message := <-h.clientMessage:
			if message.msgType == wantMatchMaking {
				if h.pendingMatch == nil {
					// if no match, just create new and add client
					h.pendingMatch = newMatch(message.sender)
				} else if !h.pendingMatch.addClient(message.sender) {
					// if client could not be added to match
					// new match and add client
					h.pendingMatch = newMatch(message.sender)
				}
			}

			match := message.sender.currentMatch
			if match != nil {
				match.clientMessage(message)
			}
		case <-h.close:
			if len(h.clients) == 0 {
				return
			}
		}

	}
}

func (h *hub) stop() {
	close(h.close)
	h.wg.Wait()
}
