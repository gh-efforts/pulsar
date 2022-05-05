# build/docker/builder.tpl
# partial for building bony without the entrypoint to allow for additional steps to be added

# ARG GO_BUILD_IMAGE is the image tag to use when building bony
ARG GO_BUILD_IMAGE
FROM $GO_BUILD_IMAGE AS builder

RUN apt-get update
RUN apt-get install -y \
  hwloc \
  jq \
  libhwloc-dev \
  mesa-opencl-icd \
  ocl-icd-opencl-dev

WORKDIR /go/src/bony
COPY . /go/src/bony

RUN make deps
RUN go mod download

# ARG BONY_NETWORK_TARGET determines which network the bony binary is built for.
# Options: mainnet, nerpanet, calibnet, butterflynet, interopnet, 2k
# See https://network.filecoin.io/ for more information about network_targets.
ARG BONY_NETWORK_TARGET=mainnet

RUN make $BONY_NETWORK_TARGET
RUN cp ./bony /usr/bin/
