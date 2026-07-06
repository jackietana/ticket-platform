FROM alpine:3.19

WORKDIR /app

COPY auth-service/auth-app .
COPY ../initdb/migrations/auth ./migrations
COPY auth-service/configs/main.yaml ./configs/main.yaml

EXPOSE 8081 9091

CMD ["./auth-app"]
