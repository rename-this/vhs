.PHONY: test
test:
	go test -cover -race ./...

.PHONY: dev
dev: 
	docker build -t vhs:dev . && \
	docker run -d -v $$(pwd):/go/vhs --name vhs_dev -it vhs:dev 