PACKAGES = $(shell go list ./... | grep -v /vendor/)

.PHONY: all
all: install

.PHONY: clean
clean:
	go clean -i ./...

.PHONY: install
install:
	go install -v

.PHONY: build
build:
	go build -v

.PHONY: fmt
fmt:
	go fmt $(PACKAGES)

.PHONY: vet
vet:
	go vet $(PACKAGES)

.PHONY: lint
lint:
	@which golint > /dev/null; if [ $$? -ne 0 ]; then \
		go get -u github.com/golang/lint/golint; \
	fi
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;
