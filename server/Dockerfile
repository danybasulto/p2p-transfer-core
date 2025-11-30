# ETAPA 1: Builder (Compilacion)
# Usamos una imagen oficial ligera para compilar
FROM golang:1.23-alpine AS builder

# Instalar certificados CA (necesarios para llamadas externas HTTPS/TLS)
# y git si fuera necesario para dependencias privadas
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app

# Primero copiamos go.mod y go.sum para aprovechar la cache de capas de Docker
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el codigo fuente
COPY . .

# Compilacion ESTATICA y OPTIMIZADA
# CGO_ENABLED=0: Deshabilita linkeo dinámico a librerias de C (crucial para scratch).
# -ldflags="-w -s": Elimina informacion de debug y tablas de simbolos (reduce tamano del binario).
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o signaling-server ./cmd/api

# ETAPA 2: Runtime (Ejecucion)
# Usamos 'scratch': una imagen vacia. Cero bloatware.
FROM scratch

# Copiar certificados de la etapa anterior (sin esto, las peticiones TLS fallan)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copiar el binario compilado
COPY --from=builder /app/signaling-server /signaling-server

# Usuario no privilegiado para seguridad (aunque scratch no tiene usuarios, es buena practica documental)
USER 1000

# Puerto expuesto (Railway inyecta la variable PORT, pero documentamos el default)
EXPOSE 8080

# Comando de entrada
ENTRYPOINT ["/signaling-server"]