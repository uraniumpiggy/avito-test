docker dev:
	docker compose up --build
docker test:
	docker compose -f docker-compose-test.yml
open db:
	docker exec -it avito_test-database-1 mysql -u user -psecret
run:
	go run cmd/main/main.go