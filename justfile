dev:
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

agent-build:
	@mkdir -p dist
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/agent ./cmd/agent
	@multipass transfer dist/agent linuxbox:/home/ubuntu/agent
	@multipass exec linuxbox -- chmod +x /home/ubuntu/agent

agent-start: agent-build
    multipass exec linuxbox -- ./agent start

agent-open:
    @ip=$(multipass info linuxbox --format json | jq -r '.info.linuxbox.ipv4[0]'); open "http://$ip:3000"
