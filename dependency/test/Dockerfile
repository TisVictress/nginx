FROM paketobuildpacks/build:base
# Should we use a paketo jammy image instead?

ARG cnb_uid=0
ARG cnb_gid=0

USER ${cnb_uid}:${cnb_gid}

COPY entrypoint /entrypoint
COPY run-test /run-test
COPY fixtures /fixtures
ENTRYPOINT ["/entrypoint"]
