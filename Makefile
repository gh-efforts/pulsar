SHELL=/usr/bin/env bash

GO_BUILD_IMAGE?=golang:1.18
PG_IMAGE?=postgres:10
COMMIT := $(shell git rev-parse --short=8 HEAD)

# GITVERSION is the nearest tag plus number of commits and short form of most recent commit since the tag, if any
GITVERSION=$(shell git describe --always --tag --dirty)

unexport GOFLAGS

CLEAN:=
BINS:=

GOFLAGS:=

.PHONY: all
all: build

## FFI

FFI_PATH:=extern/filecoin-ffi/
FFI_DEPS:=.install-filcrypto
FFI_DEPS:=$(addprefix $(FFI_PATH),$(FFI_DEPS))

$(FFI_DEPS): build/.filecoin-install ;

build/.filecoin-install: $(FFI_PATH)
	$(MAKE) -C $(FFI_PATH) $(FFI_DEPS:$(FFI_PATH)%=%)
	@touch $@

MODULES+=$(FFI_PATH)
BUILD_DEPS+=build/.filecoin-install
CLEAN+=build/.filecoin-install

ffi-version-check:
	@[[ "$$(awk '/const Version/{print $$5}' extern/filecoin-ffi/version.go)" -eq 3 ]] || (echo "FFI version mismatch, update submodules"; exit 1)
BUILD_DEPS+=ffi-version-check

.PHONY: ffi-version-check


$(MODULES): build/.update-modules ;
# dummy file that marks the last time modules were updated
build/.update-modules:
	git submodule update --init --recursive
	touch $@

CLEAN+=build/.update-modules

# tools
toolspath:=support/tools

ldflags=-X=pulsar/version.GitVersion=$(GITVERSION)
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif
GOFLAGS+=-ldflags="$(ldflags)"

.PHONY: build
build: deps pulsar

.PHONY: deps
deps: $(BUILD_DEPS)

# test starts dependencies and runs all tests
.PHONY: test
test: testfull

.PHONY: dockerup
dockerup:
	docker-compose up -d

.PHONY: dockerdown
dockerdown:
	docker-compose down


# testshort runs tests that don't require external dependencies such as postgres or redis
.PHONY: testshort
testshort:
	go test -short ./... -v

.PHONY: pulsar
pulsar:
	rm -f pulsar
	go build $(GOFLAGS) -o pulsar -mod=readonly .
BINS+=pulsar

.PHONY: clean
clean:
	rm -rf $(CLEAN) $(BINS)

.PHONY: dist-clean
dist-clean:
	git clean -xdff
	git submodule deinit --all -f

.PHONY: test-coverage
test-coverage:
	BONY_TEST_DB="postgres://postgres:password@localhost:5432/postgres?sslmode=disable" go test -coverprofile=coverage.out ./...

# tools

$(toolspath)/bin/golangci-lint: $(toolspath)/go.mod
	@mkdir -p $(dir $@)
	(cd $(toolspath); go build -tags tools -o $(@:$(toolspath)/%=%) github.com/golangci/golangci-lint/cmd/golangci-lint)

$(toolspath)/bin/gen: $(toolspath)/go.mod
	@mkdir -p $(dir $@)
	(cd $(toolspath); go build -tags tools -o $(@:$(toolspath)/%=%) github.com/filecoin-project/statediff/types/gen)

.PHONY: actors-gen
actors-gen:
	go run ./chain/actors/agen
	go fmt ./...

.PHONY: types-gen
types-gen: $(toolspath)/bin/gen
	$(toolspath)/bin/gen ./tasks/messages/types
	go fmt ./tasks/messages/types/...

.PHONY: api-gen
api-gen:
	go run ./gen/api

# dev-nets
2k: GOFLAGS+=-tags=2k
2k: build

calibnet: GOFLAGS+=-tags=calibnet
calibnet: build

nerpanet: GOFLAGS+=-tags=nerpanet
nerpanet: build

butterflynet: GOFLAGS+=-tags=butterflynet
butterflynet: build

interopnet: GOFLAGS+=-tags=interopnet
interopnet: build

# alias to match other network-specific targets
mainnet: build


# Dockerfiles

docker-files: Dockerfile Dockerfile.dev

.PHONY: Dockerfile
Dockerfile:
	@echo "Writing ./Dockerfile..."
	@cat build/docker/header.tpl \
		build/docker/builder.tpl \
		build/docker/prod_entrypoint.tpl \
		> ./Dockerfile
CLEAN+=Dockerfile

Dockerfile.dev:
	@echo "Writing ./Dockerfile.dev..."
	@cat build/docker/header.tpl \
		build/docker/builder.tpl \
		build/docker/dev_entrypoint.tpl \
		> ./Dockerfile.dev
CLEAN+=Dockerfile.dev

# Docker images

# MAINNET
.PHONY: docker-mainnet
docker-mainnet: BONY_DOCKER_FILE ?= Dockerfile
docker-mainnet: BONY_NETWORK_TARGET ?= mainnet
docker-mainnet: BONY_IMAGE_TAG ?= $(COMMIT)
docker-mainnet: docker-build-image-template

.PHONY: docker-mainnet-push
docker-mainnet-push: docker-mainnet docker-tag-and-push-template

.PHONY: docker-mainnet-dev
docker-mainnet-dev: BONY_DOCKER_FILE ?= Dockerfile.dev
docker-mainnet-dev: BONY_NETWORK_TARGET ?= mainnet
docker-mainnet-dev: BONY_IMAGE_TAG ?= $(COMMIT)-dev
docker-mainnet-dev: docker-build-image-template

.PHONY: docker-mainnet-dev-push
docker-mainnet-dev-push: docker-mainnet-dev docker-tag-and-push-template

# CALIBNET
# CALIBNET
.PHONY: docker-calibnet
docker-calibnet: BONY_DOCKER_FILE ?= Dockerfile
docker-calibnet: BONY_NETWORK_TARGET ?= calibnet
docker-calibnet: BONY_IMAGE_TAG ?= $(COMMIT)-calibnet
docker-calibnet: docker-build-image-template

.PHONY: docker-calibnet-push
docker-calibnet-push: docker-calibnet docker-tag-and-push-template

.PHONY: docker-calibnet-dev
docker-calibnet-dev: BONY_DOCKER_FILE ?= Dockerfile.dev
docker-calibnet-dev: BONY_NETWORK_TARGET ?= calibnet
docker-calibnet-dev: BONY_IMAGE_TAG ?= $(COMMIT)-calibnet-dev
docker-calibnet-dev: docker-build-image-template

.PHONY: docker-calibnet-dev-push
docker-calibnet-dev-push: docker-calibnet-dev docker-tag-and-push-template

# INTEROPNET
.PHONY: docker-interopnet
docker-interopnet: BONY_DOCKER_FILE ?= Dockerfile
docker-interopnet: BONY_NETWORK_TARGET ?= interopnet
docker-interopnet: BONY_IMAGE_TAG ?= $(COMMIT)-interopnet
docker-interopnet: docker-build-image-template

.PHONY: docker-interopnet-push
docker-interopnet-push: docker-interopnet docker-tag-and-push-template

.PHONY: docker-interopnet-dev
docker-interopnet-dev: BONY_DOCKER_FILE ?= Dockerfile.dev
docker-interopnet-dev: BONY_NETWORK_TARGET ?= interopnet
docker-interopnet-dev: BONY_IMAGE_TAG ?= $(COMMIT)-interopnet-dev
docker-interopnet-dev: docker-build-image-template

.PHONY: docker-interopnet-dev-push
docker-interopnet-dev-push: docker-interopnet-dev docker-tag-and-push-template

# BUTTERFLYNET
.PHONY: docker-butterflynet
docker-butterflynet: BONY_DOCKER_FILE ?= Dockerfile
docker-butterflynet: BONY_NETWORK_TARGET ?= butterflynet
docker-butterflynet: BONY_IMAGE_TAG ?= $(COMMIT)-butterflynet
docker-butterflynet: docker-build-image-template

.PHONY: docker-butterflynet-push
docker-butterflynet-push: docker-butterflynet docker-tag-and-push-template

.PHONY: docker-butterflynet-dev
docker-butterflynet-dev: BONY_DOCKER_FILE ?= Dockerfile.dev
docker-butterflynet-dev: BONY_NETWORK_TARGET ?= butterflynet
docker-butterflynet-dev: BONY_IMAGE_TAG ?= $(COMMIT)-butterflynet-dev
docker-butterflynet-dev: docker-build-image-template

.PHONY: docker-butterflynet-dev-push
docker-butterflynet-dev-push: docker-butterflynet-dev docker-tag-and-push-template

# NERPANET
.PHONY: docker-nerpanet
docker-nerpanet: BONY_DOCKER_FILE ?= Dockerfile
docker-nerpanet: BONY_NETWORK_TARGET ?= nerpanet
docker-nerpanet: BONY_IMAGE_TAG ?= $(COMMIT)-nerpanet
docker-nerpanet: docker-build-image-template

.PHONY: docker-nerpanet-push
docker-nerpanet-push: docker-nerpanet docker-tag-and-push-template

.PHONY: docker-nerpanet-dev
docker-nerpanet-dev: BONY_DOCKER_FILE ?= Dockerfile.dev
docker-nerpanet-dev: BONY_NETWORK_TARGET ?= nerpanet
docker-nerpanet-dev: BONY_IMAGE_TAG ?= $(COMMIT)-nerpanet-dev
docker-nerpanet-dev: docker-build-image-template

.PHONY: docker-nerpanet-dev-push
docker-nerpanet-dev-push: docker-nerpanet-dev docker-tag-and-push-template

# 2K
.PHONY: docker-2k
docker-2k: BONY_DOCKER_FILE ?= Dockerfile
docker-2k: BONY_NETWORK_TARGET ?= 2k
docker-2k: BONY_IMAGE_TAG ?= $(COMMIT)-2k
docker-2k: docker-build-image-template

.PHONY: docker-2k-push
docker-2k-push: docker-2k docker-tag-and-push-template

.PHONY: docker-2k-dev
docker-2k-dev: BONY_DOCKER_FILE ?= Dockerfile.dev
docker-2k-dev: BONY_NETWORK_TARGET ?= 2k
docker-2k-dev: BONY_IMAGE_TAG ?= $(COMMIT)-2k-dev
docker-2k-dev: docker-build-image-template

.PHONY: docker-2k-dev-push
docker-2k-dev-push: docker-2k-dev docker-tag-and-push-template


.PHONY: docker-build-image-template
docker-build-image-template:
	@echo "Building pulsar docker image for '$(BONY_NETWORK_TARGET)'..."
	docker build -f $(BONY_DOCKER_FILE) \
		--build-arg BONY_NETWORK_TARGET=$(BONY_NETWORK_TARGET) \
		--build-arg GO_BUILD_IMAGE=$(GO_BUILD_IMAGE) \
		-t $(BONY_IMAGE_NAME) \
		-t $(BONY_IMAGE_NAME):latest \
		-t $(BONY_IMAGE_NAME):$(BONY_IMAGE_TAG) \
		.

.PHONY: docker-tag-and-push-template
docker-tag-and-push-template:
	./scripts/push-docker-tags.sh $(BONY_IMAGE_NAME) $(BONY_IMAGE_TAG)

.PHONY: docker-image
docker-image: docker-mainnet
	@echo "*** Deprecated make target 'docker-image': Please use 'make docker-mainnet' instead. ***"




.PHONY: checklint
checklint:
ifeq (, $(shell which golangci-lint))
	@echo 'error: golangci-lint is not installed, please exec `brew install golangci-lint`'
	@exit 1
endif

.PHONY: lint
lint: checklint
	golangci-lint run --skip-dirs-use-default



.PHONY: generate
# generate
generate:
	go mod tidy
	go get github.com/google/wire/cmd/wire@latest
	go generate ./...

checkwire:
ifeq (,$(shell which wire))
	@echo 'error: wire is not installed, please refer to the following instructions https://github.com/google/wire#installing'
endif

wire:checkwire
	cd ${PWD}/internal && wire

.PHONY: run
run:
	go run main.go http


