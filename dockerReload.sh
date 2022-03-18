docker-compose down -v
docker-compose build --no-cache
docker-compose -f docker-compose.yaml up -d
