# Info on how to use this docker image can be found in DOCKER_README.md
ARG IMG_TAG=latest

# Compile the kiichaind binary
FROM golang:1.23-alpine AS kiichaind-builder
WORKDIR /src/app/
ENV PACKAGES="curl make git libc-dev bash file gcc linux-headers eudev-dev"
RUN apk add --no-cache $PACKAGES

# See https://github.com/CosmWasm/wasmvm/releases
ARG WASMVM_VERSION=v3.0.0
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep a04969b4f931be1d2e2f3f2313a68a20a202693f67559f7aaf1dd97250823aec
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep cefe73f0caa5a9eaba3733c639cdf5040c4746a93c20670ba6c9287fe39448ba
RUN cp "/lib/libwasmvm_muslc.$(uname -m).a" /lib/libwasmvm_muslc.a

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN LEDGER_ENABLED=false LINK_STATICALLY=true BUILD_TAGS=muslc make build
RUN echo "Ensuring binary is statically linked ..."  \
    && file /src/app/build/kiichaind | grep "statically linked"

FROM alpine:$IMG_TAG
RUN apk add --no-cache build-base jq
RUN addgroup -g 1025 nonroot
RUN adduser -D nonroot -u 1025 -G nonroot
ARG IMG_TAG
COPY --from=kiichaind-builder  /src/app/build/kiichaind /usr/local/bin/
EXPOSE 26656 26657 1317 9090 8545
USER nonroot

ENTRYPOINT ["kiichaind", "start"]
