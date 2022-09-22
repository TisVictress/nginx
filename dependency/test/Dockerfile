FROM ubuntu:18.04

# RUN apt-get -y update
# RUN apt-get -y install build-essential curl zlib1g zlib1g-dev libssl-dev libpcre3 libpcre3-dev

# ARG cnb_uid=0
# ARG cnb_gid=0

# USER ${cnb_uid}:${cnb_gid}

COPY entrypoint /entrypoint
COPY fixtures /fixtures
ENTRYPOINT ["/entrypoint"]
