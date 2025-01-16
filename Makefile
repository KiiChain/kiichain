#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --tags))
COMMIT := $(shell git log -1 --format='%H')

BUILDDIR ?= $(CURDIR)/build
INVARIANT_CHECK_INTERVAL ?= $(INVARIANT_CHECK_INTERVAL:-0)
export PROJECT_HOME=$(shell git rev-parse --show-toplevel)
export GO_PKG_PATH=$(HOME)/go/pkg
export GO111MODULE = on

# process build tags

LEDGER_ENABLED ?= true
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
	ifeq ($(OS),Windows_NT)
		GCCEXE = $(shell where gcc.exe 2> NUL)
		ifeq ($(GCCEXE),)
			$(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
		else
			build_tags += ledger
		endif
	else
		UNAME_S = $(shell uname -s)
		ifeq ($(UNAME_S),OpenBSD)
			$(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
		else
			GCC = $(shell command -v gcc 2> /dev/null)
			ifeq ($(GCC),)
				$(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
			else
				build_tags += ledger
			endif
		endif
	endif
endif

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=kiichain \
			-X github.com/cosmos/cosmos-sdk/version.ServerName=kiichaind \
			-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
			-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
			-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

# BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)' -race
BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

#### Command List ####

all: lint install

install: go.sum
		go install $(BUILD_FLAGS) ./cmd/kiichaind

install-with-race-detector: go.sum
		go install -race $(BUILD_FLAGS) ./cmd/kiichaind

# install-price-feeder: go.sum
# 		go install $(BUILD_FLAGS) ./oracle/price-feeder

loadtest: go.sum
		go build $(BUILD_FLAGS) -o ./build/loadtest ./loadtest/

# price-feeder: go.sum
# 		go build $(BUILD_FLAGS) -o ./build/price-feeder ./oracle/price-feeder

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		@go mod verify

build:
	go build $(BUILD_FLAGS) -o ./build/kiichaind ./cmd/kiichaind

# build-price-feeder:
# 	go build $(BUILD_FLAGS) -o ./build/price-feeder ./oracle/price-feeder

clean:
	rm -rf ./build

build-loadtest:
	go build -o build/loadtest ./loadtest/


###############################################################################
###                       Local testing using docker container              ###
###############################################################################
# To start a 4-node cluster from scratch:
# make clean && make docker-cluster-start
# To stop the 4-node cluster:
# make docker-cluster-stop
# If you have already built the binary, you can skip the build:
# make docker-cluster-start-skipbuild
###############################################################################


# Build linux binary on other platforms
build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-gnu-gcc make build
.PHONY: build-linux

build-price-feeder-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-gnu-gcc make build-price-feeder
.PHONY: build-price-feeder-linux

# Build docker image
build-docker-prime:
	@cd docker && docker build --tag kiichain3/prime prime --platform linux/x86_64
.PHONY: build-docker-prime

build-rpc-node:
	@cd docker && docker build --tag kiichain3/rpcnode rpcnode --platform linux/x86_64
.PHONY: build-rpc-node

# Run a single node docker container
run-prime-node: kill-kiichain-node build-docker-prime
	@rm -rf $(PROJECT_HOME)/build/generated
	docker run --rm \
	--name kiichain-node \
	-v $(PROJECT_HOME):/kiichain/kiichain3:Z \
	-v $(GO_PKG_PATH)/mod:/root/go/pkg/mod:Z \
	-v $(shell go env GOCACHE):/root/.cache/go-build:Z \
	-p 26668-26670:26656-26658 \
	-p 1317:1317 \
	--platform linux/x86_64 \
	kiichain3/prime
.PHONY: run-prime-node

# Run a single rpc state sync node docker container
run-rpc-node: build-rpc-node
	docker run --rm \
	--name kiichain-rpc-node \
	-v $(PROJECT_HOME):/kiichain/kiichain3:Z \
	-v $(PROJECT_HOME)/../sei-tendermint:/kiichain/kiichain-tendermint:Z \
    -v $(PROJECT_HOME)/../sei-cosmos:/kiichain/kiichain-cosmos:Z \
    -v $(PROJECT_HOME)/../sei-db:/kiichain/kiichain-db:Z \
	-v $(GO_PKG_PATH)/mod:/root/go/pkg/mod:Z \
	-v $(shell go env GOCACHE):/root/.cache/go-build:Z \
	-p 26668-26670:26656-26658 \
	-p 8545-8546:8545-8546 \
	-p 1317:1317 \
	--platform linux/x86_64 \
	kiichain3/rpcnode
.PHONY: run-rpc-node

run-rpc-node-skipbuild: build-rpc-node
	docker run --rm \
	--name kiichain-rpc-node \
	--network docker_localnet \
	--user="$(shell id -u):$(shell id -g)" \
	-v $(PROJECT_HOME):/kiichain/kiichain3:Z \
	-v $(PROJECT_HOME)/../sei-tendermint:/kiichain/kiichain-tendermint:Z \
    -v $(PROJECT_HOME)/../sei-cosmos:/kiichain/kiichain-cosmos:Z \
    -v $(PROJECT_HOME)/../sei-db:/kiichain/kiichain-db:Z \
	-v $(GO_PKG_PATH)/mod:/root/go/pkg/mod:Z \
	-v $(shell go env GOCACHE):/root/.cache/go-build:Z \
	-p 26668-26670:26656-26658 \
	--platform linux/x86_64 \
	--env SKIP_BUILD=true \
	kiichain3/rpcnode
.PHONY: run-rpc-node

kill-kiichain-node:
	docker ps --filter name=kiichain-node --filter status=running -aq | xargs docker kill 2> /dev/null || true

kill-rpc-node:
	docker ps --filter name=kiichain-rpc-node --filter status=running -aq | xargs docker kill 2> /dev/null || true

# Run a 4-node docker containers
docker-cluster-start: docker-cluster-stop build-docker-prime
	@rm -rf $(PROJECT_HOME)/build/generated
	@mkdir -p $(shell go env GOPATH)/pkg/mod
	@mkdir -p $(shell go env GOCACHE)
	@cd docker && PROJECT_HOME=$(PROJECT_HOME) USERID=$(shell id -u) GROUPID=$(shell id -g) GOCACHE=$(shell go env GOCACHE) NUM_ACCOUNTS=10 INVARIANT_CHECK_INTERVAL=${INVARIANT_CHECK_INTERVAL} UPGRADE_VERSION_LIST=${UPGRADE_VERSION_LIST} docker compose up

.PHONY: localnet-start

# Use this to skip the kiichaind build process
docker-cluster-start-skipbuild: docker-cluster-stop build-docker-prime
	@rm -rf $(PROJECT_HOME)/build/generated
	@cd docker && USERID=$(shell id -u) GROUPID=$(shell id -g) GOCACHE=$(shell go env GOCACHE) NUM_ACCOUNTS=10 SKIP_BUILD=true docker compose up
.PHONY: localnet-start

# Stop 4-node docker containers
docker-cluster-stop:
	@cd docker && USERID=$(shell id -u) GROUPID=$(shell id -g) GOCACHE=$(shell go env GOCACHE) docker compose down
.PHONY: localnet-stop

###############################################################################
###                        Integration Docker                               ###
###############################################################################

build-docker-integration:
	@cd integration_test/docker && docker build --tag kiichain3/integration localnode --platform linux/x86_64
.PHONY: build-docker-integration

docker-cluster-stop-integration:
	@cd integration_test/docker && USERID=$(shell id -u) GROUPID=$(shell id -g) GOCACHE=$(shell go env GOCACHE) docker compose down

docker-cluster-start-integration: docker-cluster-stop-integration build-docker-integration
	@rm -rf $(PROJECT_HOME)/build/generated
	@mkdir -p $(shell go env GOPATH)/pkg/mod
	@mkdir -p $(shell go env GOCACHE)
	@cd integration_test/docker && PROJECT_HOME=$(PROJECT_HOME) USERID=$(shell id -u) GROUPID=$(shell id -g) GOCACHE=$(shell go env GOCACHE) NUM_ACCOUNTS=10 INVARIANT_CHECK_INTERVAL=${INVARIANT_CHECK_INTERVAL} UPGRADE_VERSION_LIST=${UPGRADE_VERSION_LIST} docker compose up

build-rpc-node-integration:
	@cd integration_test/docker && docker build --tag kiichain3/rpcnode rpcnode --platform linux/x86_64
.PHONY: build-rpc-node

run-rpc-node-skipbuild-integration: build-rpc-node-integration
	docker run --rm \
	--name kii-rpc-node \
	--network docker_localnet \
	--user="$(shell id -u):$(shell id -g)" \
	-v $(PROJECT_HOME):/kiichain/kiichain3:Z \
	-v $(PROJECT_HOME)/../kii-tendermint:/kiichain/kii-tendermint:Z \
    -v $(PROJECT_HOME)/../kii-cosmos:/kiichain/kii-cosmos:Z \
    -v $(PROJECT_HOME)/../kii-db:/kiichain/kii-db:Z \
	-v $(GO_PKG_PATH)/mod:/root/go/pkg/mod:Z \
	-v $(shell go env GOCACHE):/root/.cache/go-build:Z \
	-p 26668-26670:26656-26658 \
	--platform linux/x86_64 \
	--env SKIP_BUILD=true \
	kii-chain/rpcnode

###############################################################################
###                               Tests                                     ###
###############################################################################

# Implements test splitting and running. This is pulled directly from
# the github action workflows for better local reproducibility.

GO_TEST_FILES != find $(CURDIR) -name "*_test.go"

# default to four splits by default
NUM_SPLIT ?= 4

$(BUILDDIR):
	mkdir -p $@

# The format statement filters out all packages that don't have tests.
# Note we need to check for both in-package tests (.TestGoFiles) and
# out-of-package tests (.XTestGoFiles).
$(BUILDDIR)/packages.txt:$(GO_TEST_FILES) $(BUILDDIR)
	go list -f "{{ if (or .TestGoFiles .XTestGoFiles) }}{{ .ImportPath }}{{ end }}" ./... | sort > $@

split-test-packages:$(BUILDDIR)/packages.txt
	split -d -n l/$(NUM_SPLIT) $< $<.
test-group-%:split-test-packages
	cat $(BUILDDIR)/packages.txt.$* | xargs go test -tags='norace' -p=2 -mod=readonly -timeout=10m -coverprofile=$*.profile.out -covermode=atomic
test-unit:split-test-packages
	cat $(BUILDDIR)/packages.txt | xargs go test -tags='norace' -p=2 -mod=readonly -timeout=10m -coverprofile=$*.profile.out -covermode=atomic

##############################################################################
###                                  Lint                                  ###
##############################################################################

golangci_lint_cmd=golangci-lint
golangci_version=v1.60.1
govulcheck_version=latest

lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=10m --out-format=tab
	go mod verify

vulncheck:
	GOBIN=$(BUILDDIR) go install golang.org/x/vuln/cmd/govulncheck@$(govulcheck_version)
	$(BUILDDIR)/govulncheck ./...

###############################################################################
###                       Upgrade using cosmovisor                          ###
###############################################################################

GENESIS_BIN_PATH = /root/.kiichain3/cosmovisor/genesis/bin

.PHONY: prepare-upgrade
prepare-upgrade:
	@if [ -z "$(UPGRADE_NAME)" ]; then echo "Error: UPGRADE_NAME is not set. Use \`make upgrade UPGRADE_NAME=<name>\` to specify the upgrade name."; exit 1; fi
	@if [ -z "$(CONTAINER_NAME)" ]; then echo "Error: CONTAINER_NAME is not set. Use \`make upgrade CONTAINER_NAME=<name>\` to specify the container name."; exit 1; fi
	@echo "Compiling new binary"
	@make install
	@echo "Creating the upgrade folder in cosmovisor inside the Docker container"
	@docker exec $(CONTAINER_NAME) mkdir -p /root/.kiichain3/cosmovisor/upgrades/$(UPGRADE_NAME)/bin
	@echo "Copying the binary from local GOBIN to the Docker container"
	@docker cp ~/go/bin/kiichaind $(CONTAINER_NAME):/root/.kiichain3/cosmovisor/upgrades/$(UPGRADE_NAME)/bin/kiichaind
	@docker exec $(CONTAINER_NAME) chmod +x /root/.kiichain3/cosmovisor/upgrades/$(UPGRADE_NAME)/bin/kiichaind
	@echo "Preparing for upgrade '$(UPGRADE_NAME)' completed"

.PHONY: verify
verify:
	@if [ -z "$(UPGRADE_NAME)" ]; then echo "Error: UPGRADE_NAME is not set. Use \`make upgrade UPGRADE_NAME=<name>\` to specify the upgrade name."; exit 1; fi
	@if [ -z "$(CONTAINER_NAME)" ]; then echo "Error: CONTAINER_NAME is not set. Use \`make upgrade CONTAINER_NAME=<name>\` to specify the container name."; exit 1; fi
	@echo "Checking cosmovisor settings inside the Docker container"
	@docker exec $(CONTAINER_NAME) sh -c '[ -f $(GENESIS_BIN_PATH)/kiichaind ] && echo "Cosmovisor: Genesis binary found." || (echo "Cosmovisor: Genesis binary not found." && exit 1)'
	@docker exec $(CONTAINER_NAME) sh -c '[ -d /root/.kiichain3/cosmovisor/upgrades/$(UPGRADE_NAME) ] && echo "Cosmovisor: Upgrade folder '$(UPGRADE_NAME)' found." || (echo "Cosmovisor: Upgrade folder '$(UPGRADE_NAME)' not found." && exit 1)'
	@echo "Cosmovisor settings verified successfully"

.PHONY: upgrade
upgrade:
	@if [ -z "$(UPGRADE_NAME)" ]; then echo "Error: UPGRADE_NAME is not set. Use \`make upgrade UPGRADE_NAME=<name>\` to specify the upgrade name."; exit 1; fi
	@if [ -z "$(CONTAINER_NAME)" ]; then echo "Error: CONTAINER_NAME is not set. Use \`make upgrade CONTAINER_NAME=<name>\` to specify the container name."; exit 1; fi
	@echo "Starting upgrade process for $(UPGRADE_NAME)..."
	@$(MAKE) prepare-upgrade UPGRADE_NAME=$(UPGRADE_NAME) CONTAINER_NAME=$(CONTAINER_NAME)
	@$(MAKE) verify UPGRADE_NAME=$(UPGRADE_NAME) CONTAINER_NAME=$(CONTAINER_NAME)
	@echo "Restarting docker container $(CONTAINER_NAME)..."
	@docker restart $(CONTAINER_NAME)
	@echo "Upgrade completed. The node is ready for the upgrade block."

###############################################################################
###                                Protobuf                                 ###
###############################################################################

# Proto generation uses ignite v0.23.0 (https://github.com/ignite/cli/releases/tag/v0.23.0)
proto-gen:
	@echo "Generating Protobuf files"
	ignite generate proto-go -y

swagger-gen:
	@echo "Generating Swagger files"
	etc/update-swagger-ui-statik.sh

################################################################################
###                                Contracts                                 ###
################################################################################

# Paths and Tools
CONTRACTS_SRC_DIR = contracts/src
ARTIFACTS_DIR = x/evm/artifacts
OPENZEPPELIN_PATH = contracts/lib/openzeppelin-contracts
SOLC = solc
ABIGEN = abigen

# Check if solc and abigen are installed
check-evm-tools:
	@which $(SOLC) > /dev/null || (echo "Error: solc is not installed." && exit 1)
	@which $(ABIGEN) > /dev/null || (echo "Error: abigen is not installed." && exit 1)

# Compile a single contract
define compile_evm_contract
	git submodule update --init --recursive
	$(SOLC) --overwrite @openzeppelin=$(OPENZEPPELIN_PATH) --bin -o $(ARTIFACTS_DIR)/$(1) $(CONTRACTS_SRC_DIR)/$(2)
	$(SOLC) --overwrite @openzeppelin=$(OPENZEPPELIN_PATH) --abi -o $(ARTIFACTS_DIR)/$(1) $(CONTRACTS_SRC_DIR)/$(2)
	find $(ARTIFACTS_DIR)/$(1) -type f -name "*.bin" ! -name "$(3).bin" ! -name "legacy.bin" -delete
	find $(ARTIFACTS_DIR)/$(1) -type f -name "*.abi" ! -name "$(3).abi" ! -name "legacy.abi" -delete
	$(ABIGEN) --abi=$(ARTIFACTS_DIR)/$(1)/$(3).abi --pkg=$(1) --out=$(ARTIFACTS_DIR)/$(1)/$(1).go
endef

# Compile EVM artifacts
compile-evm-cw20: check-evm-tools
	$(call compile_evm_contract,cw20,CW20ERC20Pointer.sol,CW20ERC20Pointer)

compile-evm-cw721: check-evm-tools
	$(call compile_evm_contract,cw721,CW721ERC721Pointer.sol,CW721ERC721Pointer)

compile-evm-native: check-evm-tools
	$(call compile_evm_contract,native,NativeKiiTokensERC20.sol,NativeKiiTokensERC20)

compile-evm-wkii: check-evm-tools
	$(call compile_evm_contract,wkii,WKII.sol,WKII)

# Compile all contracts
compile-evm-all: compile-evm-cw20 compile-evm-cw721 compile-evm-native compile-evm-wkii
	@echo "All contracts compiled successfully."

.PHONY: check-evm-tools compile-evm-cw20 compile-evm-cw721 compile-evm-native compile-evm-wkii compile-evm-all
