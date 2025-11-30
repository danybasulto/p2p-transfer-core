package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Tiempo permitido para escribir un mensaje al par.
	writeWait = 10 * time.Second

	// Tiempo permitido para leer el siguiente mensaje pong del par.
	pongWait = 60 * time.Second

	// Enviar pings al par con este periodo. Debe ser menor que pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Tamano maximo permitido del mensaje (en bytes).
	// 512 bytes es suficiente para senales SDP y candidatos ICE.
	maxMessageSize = 512
)

// Upgrader convierte la conexion HTTP en una conexion WebSocket.
// CheckOrigin en true permite conexiones desde cualquier origen (CORS).
// IMPORTANTE: En produccion, esto debe restringirse al dominio de tu frontend.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client es un intermediario entre la conexion websocket y el hub.
type Client struct {
	hub *Hub

	// La conexion websocket.
	conn *websocket.Conn

	// Canal bufferizado para mensajes salientes.
	send chan []byte
}

// readPump bombea mensajes desde la conexión websocket al hub.
//
// La aplicacion se ejecuta en una goroutine por conexion. La aplicacion
// asegura que haya como maximo un lector en una conexion ejecutando todos
// los reads en esta goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	// Configuracion de limites y tiempos para evitar ataques DoS y conexiones zombies.
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { 
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil 
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// Enviamos el mensaje limpio al Hub para que decida que hacer.
		// En el futuro, aqui parsearemos el JSON para ver a que sala va.
		c.hub.broadcast <- message
	}
}

// writePump bombea mensajes del hub a la conexion websocket.
//
// Se ejecuta una goroutine por conexion para asegurar que haya como maximo un escritor.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// El hub cerro el canal.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Si hay varios mensajes en cola, los anadimos al mismo frame para optimizar red (batching).
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// Heartbeat: Mantiene la conexion viva a traves de NATs y Load Balancers.
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs maneja las solicitudes websocket del Peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Permitimos que la coleccion de memoria haga su trabajo iniciando goroutines.
	go client.writePump()
	go client.readPump()
}