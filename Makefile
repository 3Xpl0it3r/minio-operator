all: build

build:
	docker buildx build . -f build/Dockerfile --platform linux/amd64,linux/arm64 -t registry.bizsaas.net/operator/minio-operator:2022-12-06_v6 --no-cache
