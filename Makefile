.PHONY: test
test:
	go test -cover -race -coverprofile coverage.out `go list ./... | grep -v /cmd/vhs`

.PHONY: test-all
test-all: test
	go test -cover -coverprofile coverage_cmd_vhs.out ./cmd/vhs

.PHONY: dev
dev:
	docker build -t vhs:dev . && \
	docker run -d -v $$(pwd):/go/vhs -p 8888:8888 -v $$HOME/.config/gcloud:/root/.config/gcloud --name vhs_dev -it vhs:dev

VERSION ?= dev

.PHONY: docker-vhs
docker-vhs:
	docker build \
	-f ./docker/vhs/Dockerfile \
	-t ghcr.io/rename-this/vhs:$(VERSION) \
	-t ghcr.io/rename-this/vhs:$$(git rev-parse --short HEAD) \
	-t ghcr.io/rename-this/vhs:latest \
	.

.PHONY: docker-vhs-push
docker-vhs-push: docker-vhs
	docker push ghcr.io/rename-this/vhs:$(VERSION) && \
	docker push ghcr.io/rename-this/vhs:$$(git rev-parse --short HEAD) && \
	docker push ghcr.io/rename-this/vhs:latest
