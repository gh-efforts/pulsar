GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

FFI_DEPS:=.install-filcrypto
FFI_DEPS:=$(addprefix $(FFI_PATH),$(FFI_DEPS))
BUILD_DEPS+=build/.filecoin-install
$(FFI_DEPS): build/.filecoin-install ;

ffi-version-check:
	@[[ "$$(awk '/const Version/{print $$5}' extern/filecoin-ffi/version.go)" -eq 3 ]] || (echo "FFI version mismatch, update submodules"; exit 1)
BUILD_DEPS+=ffi-version-check

build/.filecoin-install: $(FFI_PATH)
	$(MAKE) -C $(FFI_PATH) $(FFI_DEPS:$(FFI_PATH)%=%)
	@touch $@

MODULES+=$(FFI_PATH)
BUILD_DEPS+=build/.filecoin-install
CLEAN+=build/.filecoin-install




checklint:
ifeq (, $(shell which golangci-lint))
	@echo 'error: golangci-lint is not installed, please exec `brew install golangci-lint`'
	@exit 1
endif

lint: checklint
	golangci-lint run --skip-dirs-use-default


.PHONY: deps
deps: $(BUILD_DEPS)



