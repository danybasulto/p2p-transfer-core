package websocket

// Hub mantiene el conjunto de clientes activos y transmite mensajes a los clientes.
type Hub struct {
	// Clientes registrados.
	clients map[*Client]bool

	// Mensajes entrantes de los clientes (broadcast a todos por ahora).
	broadcast chan []byte

	// Solicitudes de registro desde los clientes.
	register chan *Client

	// Solicitudes de cancelacion de registro desde los clientes.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			
		case message := <-h.broadcast:
			// En esta fase inicial, hacemos eco del mensaje a TODOS los conectados.
			// FASE SIGUIENTE: Aqui implementaremos la logica de "Salas" (Rooms) con Redis
			// para enviar el mensaje SOLO al peer correspondiente.
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Si el buffer de envio del cliente esta lleno, asumimos que esta muerto o es lento.
					// Lo desconectamos para no bloquear al resto del Hub.
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}