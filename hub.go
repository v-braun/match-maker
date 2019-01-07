package main

type hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	clientMessage chan *Message

	// register new client in the hub
	register chan *Client

	// unregister client in the hub
	unregister chan *Client

	end chan interface{}
}

func startHub() *hub {
	h := &hub{
		clientMessage: make(chan *Message),
		unregister:    make(chan *Client),
		register:      make(chan *Client),
		clients:       make(map[*Client]bool),
	}

	go h.run()

	return h
}

func (h *hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case <-h.end:
			h.disconnectClients()
		case message := <-h.clientMessage:
			h.handleClientMessage(message)
		}
	}
}

func (h *hub) handleClientMessage(msg *Message) {

}

func (h *hub) disconnectClients() {
	for client := range h.clients {
		close(client.send)
		client.WaitClosed()
		delete(h.clients, client)
	}
}
