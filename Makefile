SHELL := /bin/sh

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: dev db gen test build initdata

dev: db
	trap 'kill 0' EXIT; \
	(cd api && go run .) & \
	(cd web && npm run dev); \
	wait

db:
	docker compose up -d --wait

gen:
	cd api && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config oapi-codegen.yaml openapi.yaml
	cd web && npm run gen

test:
	cd api && go test ./...

build:
	rm -rf api/webdist/assets
	cd web && npm ci && npm run build
	cd api && go build -o piggy .

initdata:
	cd api && go run ./cmd/initdata
