.PHONY: test
test:
	go test -cover -race ./...

.PHONY: dev
dev: 
	docker build -t vhs:dev . && \
	docker run -d -v $$(pwd):/go/vhs -p 8888:8888 -v $$HOME/.config/gcloud:/root/.config/gcloud --name vhs_dev -it vhs:dev 
