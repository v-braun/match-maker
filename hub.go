package main

import "sync"

type hub struct {
	wg sync.WaitGroup

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	clientMessage chan *Message

	// register new client in the hub
	register chan *Client

	// unregister client in the hub
	unregister chan *Client

	// client search for a match
	enterMatch chan *Client

	// client leaves a match
	leaveMatch chan *Client

	// a message was send in a match
	matchMessage chan *Message

	pendingMatch *Match
}

func startHub() *hub {
	h := &hub{
		clientMessage: make(chan *Message),
		unregister:    make(chan *Client),
		register:      make(chan *Client),
		clients:       make(map[*Client]bool),
		enterMatch:    make(chan *Client),
		leaveMatch:    make(chan *Client),
		matchMessage:  make(chan *Message),
	}

	go h.runClientIO()
	go h.runMatchMaking()

	return h
}

func (h *hub) runMatchMaking() {
	for {
		select {
		case client := <-h.enterMatch:
			if h.pendingMatch == nil {
				// if no match, just create new and add client
				h.pendingMatch = NewMatch(client)
			} else if !h.pendingMatch.addClient(client) {
				// if client could not be added to match
				// new match and add client
				h.pendingMatch = NewMatch(client)
			}
		case client := <-h.leaveMatch:
			// if the match is ended (no players) and the pending match is that actual
			// set pending to nil
			if match := client.currentMatch; match != nil && match.removeClient(client) && h.pendingMatch == match {
				h.pendingMatch = nil
			}
		case msg := <-h.matchMessage:
			if match := msg.sender.currentMatch; match != nil {
				match.clientMessage(msg)
			}
		}

	}
}

func (h *hub) runClientIO() {
	defer func() {
		close(h.leaveMatch)
		close(h.enterMatch)
		close(h.matchMessage)
	}()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.wg.Add(1)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if client.currentMatch != nil {
					h.leaveMatch <- client
				}
				h.wg.Done()
			}
		case message := <-h.clientMessage:
			go h.handleClientMessage(message)
		}
	}
}

func (h *hub) handleClientMessage(msg *Message) {
	c := msg.sender
	switch msg.msgType {
	case WantMatchMaking:
		h.enterMatch <- c
	case MatchMessage:
		//handleMatchMessage(msg)
	case WantLeaveMatch:
		h.leaveMatch <- c
	}
}

func stop() {

}
