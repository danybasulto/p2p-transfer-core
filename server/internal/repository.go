package internal

import "context"

// RoomRepository define las operaciones que necesitamos para gestionar salas.
// Usamos interfaces para desacoplar la logica de negocio de la base de datos (DIP).
type RoomRepository interface {
    // CreateRoom guarda una sala y retorna error si falla.
    CreateRoom(ctx context.Context, roomID string) error
    
    // AddPeerToRoom registra un usuario en una sala existente.
    AddPeerToRoom(ctx context.Context, roomID string, peerID string) error
    
    // RoomExists verifica si la sala esta activa.
    RoomExists(ctx context.Context, roomID string) (bool, error)
}