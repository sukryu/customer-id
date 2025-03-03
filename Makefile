.PHONY: test-unit
test-unit:
	go test -v ./tests/unit/...

.PHONY: gen-keys
gen-keys:
	@mkdir -p internal/config/keys/$(ENV)
	openssl genrsa -out internal/config/keys/$(ENV)/private.pem 2048
	openssl rsa -in internal/config/keys/$(ENV)/private.pem -pubout -out internal/config/keys/$(ENV)/public.pem

.PHONY: docker-up docker-down

docker-up:
	docker-compose -f deploy/docker/docker-compose.yml up -d
	@echo "Waiting for services to be healthy..."
	@sleep 10  # Give services time to start
	@docker-compose -f deploy/docker/docker-compose.yml ps

docker-down:
	docker-compose -f deploy/docker/docker-compose.yml down -v