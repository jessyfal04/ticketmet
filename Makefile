.PHONY: test server docker-build docker-push docker-deploy docker-up docker-down docker-logs

IMAGE ?= docker.io/jessyfal04/ticketmet:latest

-include .secrets/ticketmaster.mk

export TICKETMASTER_API_KEY
export TICKETMASTER_CONSUMER_SECRET

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
	@test -n "$$TICKETMASTER_API_KEY" || { echo "TICKETMASTER_API_KEY missing. Run: TICKETMASTER_API_KEY=... make docker-deploy"; exit 1; }
	ssh vps "\
		set -e; \
		mkdir -p /opt/ticketmet/data; \
		docker pull $(IMAGE); \
		docker run --rm --user root -v /opt/ticketmet/data:/data --entrypoint chown $(IMAGE) -R ticketmet:ticketmet /data; \
		docker rm -f ticketmet 2>/dev/null || true; \
		docker run -d --name ticketmet --restart unless-stopped -p 127.0.0.1:11200:8080 -e TICKETMASTER_API_KEY='$$TICKETMASTER_API_KEY' -v /opt/ticketmet/data:/app/data $(IMAGE); \
		docker ps --filter name=ticketmet"
