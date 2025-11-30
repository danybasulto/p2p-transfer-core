package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRepository implementa la interfaz internal.RoomRepository
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository inicializa la conexion.
// Recibe la URL de conexion de Upstash.
func NewRedisRepository(connectionString string) (*RedisRepository, error) {
	opts, err := redis.ParseURL(connectionString)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	client := redis.NewClient(opts)

	// Verificamos conexion inmediatamente (Fail Fast)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisRepository{client: client}, nil
}

// TTL de la sala: 24 horas. Si nadie la usa, se borra sola.
const roomTTL = 24 * time.Hour

func (r *RedisRepository) CreateRoom(ctx context.Context, roomID string) error {
	// Usamos SETNX (Set if Not Exists) para evitar sobrescribir salas.
	// Clave: "room:{id}", Valor: "created" (o timestamp)
	key := fmt.Sprintf("room:%s", roomID)
	
	success, err := r.client.SetNX(ctx, key, time.Now().String(), roomTTL).Result()
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("room already exists")
	}
	return nil
}

func (r *RedisRepository) RoomExists(ctx context.Context, roomID string) (bool, error) {
	key := fmt.Sprintf("room:%s", roomID)
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddPeerToRoom en esta fase simplificada solo valida que la sala exista.
// En WebRTC puro, el mapeo de peers suele estar en memoria del Websocket Hub,
// Redis se usa mas para validar que el "ID de la sala" es legitimo.
func (r *RedisRepository) AddPeerToRoom(ctx context.Context, roomID string, peerID string) error {
	exists, err := r.RoomExists(ctx, roomID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("room not found")
	}
	// Aqui podriamos anadir logica para contar participantes si quisieramos limitar a 2.
	return nil
}