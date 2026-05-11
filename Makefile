.PHONY: test server

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
