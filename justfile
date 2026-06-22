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

agent-version := `git describe --tags --always --dirty 2>/dev/null || echo dev`

agent-build-linux:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-X main.version={{agent-version}}" -o dist/agent-linux-arm64/syslantern ./cmd/agent

agent-package: agent-build-linux
	# COPYFILE_DISABLE stops macOS tar from adding AppleDouble ._* metadata files.
	# --no-xattrs stops macOS tar from adding LIBARCHIVE.xattr.* headers that Linux tar warns about.
	COPYFILE_DISABLE=1 tar --no-xattrs -czf public/syslantern-agent.tar.gz -C dist/agent-linux-arm64 syslantern

agent-install: agent-build-linux
	multipass transfer dist/agent-linux-arm64/syslantern linuxbox:/tmp/syslantern-agent
	multipass transfer scripts/install-multipass-agent.sh linuxbox:/tmp/install-multipass-agent.sh
	multipass exec linuxbox -- chmod +x /tmp/install-multipass-agent.sh
	multipass exec linuxbox -- sudo /tmp/install-multipass-agent.sh
