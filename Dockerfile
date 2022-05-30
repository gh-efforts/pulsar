FROM golang:1.18.0 AS builder


RUN apt-get update && apt-get install -y ca-certificates build-essential clang ocl-icd-opencl-dev ocl-icd-libopencl1 jq libhwloc-dev


#RUN apt-get update
#RUN apt-get install -y \
#  hwloc \
#  jq \
#  libhwloc-dev \
#  mesa-opencl-icd \
#  ocl-icd-opencl-dev

WORKDIR /go/src/pulsar
COPY . /go/src/pulsar

# RUN make clean deps
RUN make deps
RUN GOPROXY=https://goproxy.cn


ENV LIBRARY_PATH /opt/homebrew/lib
ENV FFI_BUILD_FROM_SOURCE 1
ENV PATH ="(brew --prefix coreutils)/libexec/gnubin:/usr/local/bin:$PATH"

ARG BONY_NETWORK_TARGET=mainnet

RUN make $BONY_NETWORK_TARGET
RUN cp ./pulsar /usr/bin/



FROM buildpack-deps:bullseye-curl


RUN apt-get update
RUN apt-get install -y --no-install-recommends \
    jq


COPY --from=builder /go/src/pulsar/pulsar /usr/bin/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libOpenCL.so* /lib/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libhwloc.so* /lib/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libnuma.so* /lib/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libltdl.so* /lib/

EXPOSE 8078

ENTRYPOINT ["/usr/bin/pulsar"]
CMD ["--help"]