FROM alpine:3.19

WORKDIR /app

COPY catalog-service/catalog-app .
COPY ../initdb/migrations/catalog ./migrations
COPY catalog-service/configs/main.yaml ./configs/main.yaml

EXPOSE 8082 9092

CMD ["./catalog-app"]
