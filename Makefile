postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=Ilkin561 -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:Ilkin561@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:Ilkin561@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	docker run --rm -v "C:\Users\KBI-07\Desktop\simplebank:/src" -w /src kjconroy/sqlc init

test:
	go test -v -cover ./...
.PHONY:createdb	postgres dropdb migrateup migratedown 