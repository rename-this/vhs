.PHONY: test
test:
	go test -cover -race ./...

.PHONY: dev
dev: 
	docker build -t vhs:dev . && \
	docker run -v $$(pwd):/go/vhs -it vhs:dev