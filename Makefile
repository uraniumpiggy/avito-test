docker dev:
	docker compose up --build
open db:
	docker exec -it avito_test-database-1 mysql -u user -psecret
run:
	go run cmd/main/main.go