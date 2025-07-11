docker-compose down smtp-relay-api
docker-compose up -d --build smtp-relay-api
docker-compose logs -f smtp-relay-api