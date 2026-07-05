FROM alpine:3.19

WORKDIR /app

COPY catalog-service/catalog-app .
COPY catalog-service/migrations ./migrations
COPY catalog-service/configs/main.yml ./configs/main.yml

EXPOSE 8080 9092

CMD ["./catalog-app"]
