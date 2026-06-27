.PHONY: install run test test-go test-python lint format build verify

install:
	python -m pip install -r requirements.txt
	python -m pip install -e ./python

run:
	cd go && go run ./cmd/gateway

test: test-go test-python

test-go:
	cd go && go test -buildvcs=false ./...

test-python:
	python -m pytest python/tests -q

lint:
	cd go && go vet ./...
	ruff check python

format:
	cd go && gofmt -w .
	ruff format python

build:
	mkdir -p bin
	cd go && for cmd in gateway operator trace-proxy feature-gateway storage-proxy metrics-collector cli; do \
		go build -buildvcs=false -o ../bin/mlaiops-$$cmd ./cmd/$$cmd; \
	done

verify: test lint build
	node --check go/cmd/gateway/web/app.js
	python -m compileall -q python/mlaiops_sdk
	! rg -i '\b(mlrun|nuclio|v3io|iguazio)\b' go config
