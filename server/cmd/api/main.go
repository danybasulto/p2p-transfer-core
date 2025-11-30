package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/danybasulto/p2p-signaling/internal/websocket"
)

var addr = flag.String("addr", ":8080", "dirección de servicio http")

func main() {
	flag.Parse()

	// Inicializamos el Hub de WebSockets
	hub := websocket.NewHub()
	// Corremos el Hub en su propia goroutine (background)
	go hub.Run()

	// Definimos la ruta para los WebSockets
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// Health check simple (útil para que Railway sepa que estamos vivos)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Servidor de Señalización iniciando en %s", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}