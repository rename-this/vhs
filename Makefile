.PHONY: test
test:
	docker build \
		--target base \
		--tag vhs:test \
		--file docker/vhs/Dockerfile \
		. && \
	docker run \
		--rm \
		--volume /var/run/docker.sock:/var/run/docker.sock \
		--network host \
		-i vhs:test \
		go test -cover -race -coverprofile coverage.out `go list ./... | grep -v -f .testignore`

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

docs: submods
	hugo serve

docs-ci: init submods
	hugo -s site -b / --gc

init: submods
	npm install -D --save postcss postcss-cli autoprefixer

submods:
	git submodule update --init --recursive
