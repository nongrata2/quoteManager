# Getting started:
1. Create and fill .env file in root folder. Example: 
```
HTTP_SERVER_ADDRESS=:8080
HTTP_SERVER_TIMEOUT=5s
LOG_LEVEL=DEBUG
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=postgres
DB_PORT=5432
```
2. Start the program with command:
```sh
docker compose up --build
```

You can run tests from root directory with: 
```sh
go test ./... -v
```

# Example of commands:
1. Create Quote:
```sh
curl -X POST -H "Content-Type: application/json" -d '{"author":"Confucius", "quote":"Life is simple, but we insist on making it complicated."}' http://localhost:8081/quotes
```
2. Get all the Quotes:
```sh
curl -X GET http://localhost:8081/quotes
```
3. Get the Quotes with filter on authors:
```sh
curl http://localhost:8081/quotes?author=Confucius
```
4. Get random Quote:
```sh
curl http://localhost:8081/quotes/random
```
5. Delete the Quote with ID:
```sh
curl -X DELETE http://localhost:8081/quotes/{quoteID}
```

# Versions:
- Golang 1.23.6
- Docker 26.1.3
- PostgreSQL 16.9