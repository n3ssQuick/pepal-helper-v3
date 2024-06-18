FROM golang:latest as build

WORKDIR /app

COPY . .
COPY .env .env

RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o helper-api .

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/helper-api /app/helper-api
COPY .env /app/.env

EXPOSE 8888

RUN chmod +x /app/helper-api

ENTRYPOINT [ "/app/helper-api" ]
