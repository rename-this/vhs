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
	-t us.gcr.io/carbon-relay-dev/vhs:$(VERSION) \
	-t us.gcr.io/carbon-relay-dev/vhs:$$(git rev-parse --short HEAD) \
	-t us.gcr.io/carbon-relay-dev/vhs:latest \
	.

.PHONY: docker-vhs-push
docker-vhs-push: docker-vhs
	docker push us.gcr.io/carbon-relay-dev/vhs:$(VERSION) && \
	docker push us.gcr.io/carbon-relay-dev/vhs:$$(git rev-parse --short HEAD) && \
	docker push us.gcr.io/carbon-relay-dev/vhs:latest
