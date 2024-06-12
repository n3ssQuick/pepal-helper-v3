FROM golang:latest as build
WORKDIR /app
COPY . .
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o helper-api .
FROM alpine:latest
COPY --from=build /app/helper-api .
EXPOSE 8888
RUN chmod +x helper-api
ENTRYPOINT [ "/helper-api" ]