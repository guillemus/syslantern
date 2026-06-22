dev:
    bunx concurrently -k \
      -n go,css,js \
      -c green,cyan,magenta \
      "air" \
      "bunx @tailwindcss/cli -i ./views/styles.css -o ./public/styles.css --watch" \
      "bunx esbuild ./views/scripts/scripts.ts --bundle --minify --outfile=./public/scripts.js --watch"

build-assets:
	bunx @tailwindcss/cli -i ./views/styles.css -o ./public/styles.css --minify
	bunx esbuild ./views/scripts/scripts.ts --bundle --minify --outfile=./public/scripts.js

sqlc:
    sqlc generate

typecheck:
	bunx tsc --noEmit

check-schema:
	atlas schema validate \
	  --url "file://db/schema.sql" \
	  --dev-url "sqlite://dev?mode=memory"

migrate:
	atlas schema apply \
	  --to "file://db/schema.sql" \
	  --url "sqlite://tmp/syslantern.db" \
	  --dev-url "sqlite://dev?mode=memory" \
	  --auto-approve

build-linux: build-assets
	mkdir -p dist
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/syslantern ./cmd/server

agent-build:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/syslantern-agent ./cmd/agent

release-check:
	goreleaser check

agent-install: agent-build
	multipass transfer dist/syslantern-agent linuxbox:/tmp/syslantern-agent
	multipass transfer scripts/install-multipass-agent.sh linuxbox:/tmp/install-multipass-agent.sh
	multipass exec linuxbox -- chmod +x /tmp/install-multipass-agent.sh
	multipass exec linuxbox -- sudo /tmp/install-multipass-agent.sh
