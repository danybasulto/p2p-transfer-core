package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/danybasulto/p2p-signaling/internal/platform/redis"
	"github.com/danybasulto/p2p-signaling/internal/websocket"
	"github.com/joho/godotenv"
)

var addr = flag.String("addr", ":8080", "dirección de servicio http")

func main() {
	flag.Parse()

	// 1. Cargar variables de entorno (solo en local)
	// En produccion (Railway), las variables vienen del sistema, por lo que no es error si falta el archivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("Info: No se encontró archivo .env (usando variables de entorno del sistema)")
	}

	// 2. Obtener URL de Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("Error: La variable de entorno REDIS_URL es obligatoria")
	}

	// 3. Inicializar conexion a Redis
	// Esto fallara (panic/log.Fatal) si Upstash no responde, lo cual es BUENO (Fail Fast).
	repo, err := redis.NewRedisRepository(redisURL)
	if err != nil {
		log.Fatal("Error conectando a Redis: ", err)
	}
	log.Println("Conexión a Redis exitosa")

	// 4. Inyectar el repositorio al Hub
	hub := websocket.NewHub(repo)
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Servidor de Señalización iniciando en %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}