dev:
    lsof -ti :3000 | xargs kill
    bunx concurrently -k -n go,css "air" "bunx @tailwindcss/cli -i ./views/styles.css -o ./public/styles.css --watch"

build-assets:
	bunx @tailwindcss/cli -i ./views/styles.css -o ./public/styles.css --minify

sqlc:
    sqlc generate

check-schema:
	atlas schema validate \
	  --url "file://db/schema.sql" \
	  --dev-url "sqlite://dev?mode=memory"

migrate:
	atlas schema apply \
	  --to "file://db/schema.sql" \
	  --url "sqlite://tmp/app.db" \
	  --dev-url "sqlite://dev?mode=memory" \
	  --auto-approve

build-linux: build-assets
	mkdir -p dist
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/app ./cmd/server

build-agent:
	@mkdir -p dist
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/agent ./cmd/agent
	@multipass transfer dist/agent linuxbox:/home/ubuntu/agent
	@multipass exec linuxbox -- chmod +x /home/ubuntu/agent

start-agent: build-agent
    multipass exec linuxbox -- ./agent start
