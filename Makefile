.PHONY: test server docker-build docker-push docker-deploy docker-up docker-down docker-logs

IMAGE ?= docker.io/jessyfal04/ticketmet:latest
VPS ?= vps
VPS_PORT ?= 11200

TEST_SCRIPTS := $(wildcard server/test/*.sh)

test:
	@if [ -z "$(TEST_SCRIPTS)" ]; then \
		echo "No test scripts found in server/test"; \
		exit 0; \
	fi; \
	for f in $(TEST_SCRIPTS); do \
		echo ">> $$f"; \
		bash "$$f"; \
	done

server:
	go -C server run ./main

docker-build:
	docker build -t $(IMAGE) .

docker-push: docker-build
	docker push $(IMAGE)

docker-deploy:
	ssh $(VPS) 'docker pull $(IMAGE); docker stop ticketmet 2>/dev/null || true; docker rm ticketmet 2>/dev/null || true; docker run -d --name ticketmet --restart unless-stopped -p 127.0.0.1:$(VPS_PORT):8080 $(IMAGE); docker ps --filter name=ticketmet'
