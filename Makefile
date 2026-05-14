.PHONY: test server docker-build docker-push docker-deploy deploy-check

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

docker-deploy: docker-push
	@test -n "$$TICKETMASTER_API_KEY" || { echo "TICKETMASTER_API_KEY missing. Run: TICKETMASTER_API_KEY=... make docker-deploy"; exit 1; }
	ssh vps "\
		set -e; \
		mkdir -p /opt/ticketmet/data; \
		docker pull $(IMAGE); \
		docker rm -f ticketmet 2>/dev/null || true; \
		docker run --rm --user root -v /opt/ticketmet/data:/data --entrypoint rm $(IMAGE) -f /data/ticketmet.sqlite3 /data/ticketmet.sqlite3-shm /data/ticketmet.sqlite3-wal; \
		docker run --rm --user root -v /opt/ticketmet/data:/data --entrypoint chown $(IMAGE) -R ticketmet:ticketmet /data; \
		docker run -d --name ticketmet --restart unless-stopped -p 127.0.0.1:11200:8080 -e TICKETMASTER_API_KEY='$$TICKETMASTER_API_KEY' -e WEBAUTHN_RP_ID=ticketmet.jessyfal04.dev -e WEBAUTHN_ORIGINS=https://ticketmet.jessyfal04.dev -v /opt/ticketmet/data:/app/data $(IMAGE); \
		docker ps --filter name=ticketmet; \
		for i in 1 2 3 4 5 6 7 8 9 10; do \
			if wget -qO- http://127.0.0.1:11200/healthz >/dev/null; then exit 0; fi; \
			sleep 1; \
		done; \
		docker logs --tail=80 ticketmet; \
		exit 1"

deploy-check:
	ssh vps "docker ps --filter name=ticketmet && docker logs --tail=80 ticketmet && wget -qO- http://127.0.0.1:11200/healthz"
