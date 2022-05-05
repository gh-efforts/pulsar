# build/docker/dev_entrypoint.tpl
# partial for completing a dev bony dockerfile

ENTRYPOINT ["/usr/bin/bony"]
CMD ["--help"]
