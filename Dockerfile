# Stage 1: Build frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.26-alpine AS backend-builder
RUN apk add --no-cache build-base
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app
COPY backend-go/go.mod backend-go/go.sum ./
RUN go mod download
COPY backend-go/ ./
RUN go build -o server cmd/server/main.go

# Stage 3: Runtime
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend-builder /app/server .
COPY --from=backend-builder /app/.env.enc .
COPY --from=frontend-builder /app/dist ./dist
ENV PORT=8000
ENV FRONTEND_DIST=/app/dist
EXPOSE 8000
CMD ["./server"]
