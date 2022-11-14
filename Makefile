hi:
	echo "up containers & start dev env..."
run_dev: cmd/main/main.go
	nodemon --exec APP_ENV=DEV go run cmd/main/main.go --signal SIGTERM || exit 1
build: cmd/main/main.go
	go build -o build/blog cmd/main/main.go
run_build:
	APP_ENV=PROD build/./blog
up:
	docker compose up -d
down:
	docker compose down
dev: hi up run_dev
