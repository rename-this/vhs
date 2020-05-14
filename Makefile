.PHONY: test
test:
	go test -cover -race ./...

.PHONY: run
run: 
	docker build -t vhs:dev . && docker run -it vhs:dev