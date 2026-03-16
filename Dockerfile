# Stage 1: Build React frontend
FROM node:25-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY --from=frontend /app/static ./static/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o firebox-auditor .

# Stage 3: Minimal runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /app/firebox-auditor .
EXPOSE 8443
ENTRYPOINT ["./firebox-auditor"]
