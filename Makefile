build:
	go build -o lokix cmd/lokix/main.go

compose:
	docker compose up --build --abort-on-container-exit

# TODO: Add `test` as well
# TODO: Do I replace make build with docker? IDK lol