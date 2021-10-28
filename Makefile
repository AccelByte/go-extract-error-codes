lint:
	make lint-golang-ci
	make lint-outdated-dependencies

lint-outdated-dependencies:
	@echo "linting golang dependencies ..."
	@go list -u -m -json all 2> /dev/null | \
	docker run -i psampaz/go-mod-outdated:v0.7.0 -update -direct | \
	grep "github.com/AccelByte" && exit 1 || echo "OK"

lint-golang-ci:
	docker run --rm -v $(PWD):$(PWD) -w $(PWD) -u `id -u $(USER)` -e GOLANGCI_LINT_CACHE=/tmp/.cache -e GOCACHE=/tmp/.cache golangci/golangci-lint:v1.41.0 golangci-lint run -v --fix
