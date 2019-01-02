VERSION := $(shell grep -Eo '(v[0-9]+[\.][0-9]+[\.][0-9]+(-[a-zA-Z0-9]*)?)' version.go)

.PHONY: build docker release

build:
	go fmt ./...
	@mkdir -p ./bin/
	CGO_ENABLED=1 go build -o ./bin/auth github.com/moov-io/auth

docker:
	docker build --pull -t moov/auth:$(VERSION) -f Dockerfile .
	docker tag moov/auth:$(VERSION) moov/auth:latest

release: docker AUTHORS
	go test ./...
	git tag -f $(VERSION)

release-push:
	docker push moov/auth:$(VERSION)
	git push origin master
	git push --tags origin $(VERSION)

# From https://github.com/genuinetools/img
.PHONY: AUTHORS
AUTHORS:
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@echo "$(shell git log --format='\n%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf)" >> $@
