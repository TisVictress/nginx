.PHONY: test retrieve

id:
	@echo nginx

test:
	@cd test; \
	TARBALL_PATH=$(tarballPath) VERSION=$(version) go test

retrieve:
	@cd retrieval; \
	go build -o retrieve; \
	./retrieve --buildpack_toml_path=$(buildpackTomlPath) \
			--metadata_file=$(output) --targets_file=/tmp/targets.json; \
	rm retrieve