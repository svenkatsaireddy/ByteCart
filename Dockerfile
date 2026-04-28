FROM node:24-alpine AS frontend

WORKDIR /src/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend ./
RUN npm run build

FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=build /out/server /app/server
COPY --from=frontend /src/frontend/dist /app/frontend/dist
COPY data /app/data
COPY examples /app/examples
COPY api /app/api

EXPOSE 8080

ENV ADDR=:8080

CMD ["/app/server"]
