FROM alpine:3.19

WORKDIR /app

COPY auth-service/auth-app .
COPY auth-service/migrations ./migrations
COPY auth-service/configs/main.yml ./configs/main.yml

EXPOSE 8080 9000

CMD ["./auth-app"]
