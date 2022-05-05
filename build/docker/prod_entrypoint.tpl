# build/docker/prod_entrypoint.tpl
# partial for producing a minimal image by extracting the binary
# from a prior build step (builder.tpl)

FROM buildpack-deps:bullseye-curl

RUN apt-get update
RUN apt-get install -y --no-install-recommends \
      jq

COPY --from=builder /go/src/bony/bony /usr/bin/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libOpenCL.so* /lib/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libhwloc.so* /lib/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libnuma.so* /lib/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libltdl.so* /lib/

ENTRYPOINT ["/usr/bin/bony"]
CMD ["--help"]
