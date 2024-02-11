start:
		docker compose up -d && docker exec -it auth-proxy sh

stop:
		docker compose stop

attach:
		docker exec -it auth-proxy sh

remove:
		docker compose down

build:
		go build ./cmd/app -o /bin

run:
		go run ./cmd/app