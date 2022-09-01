#!/usr/bin/env bash

set -eu
set -o pipefail

WORKING_DIR=$(mktemp -d)
DEST_DIR=$(mktemp -d)
OUTPUT_DIR=$(mktemp -d)

function scream {
  echo "--- --- --- ---"
  echo "--- --- --- ---"
  echo "--- --- --- ---"
  echo "${0}"
  echo "--- --- --- ---"
  echo "--- --- --- ---"
  echo "--- --- --- ---"
}

scream "Running apt-get"

sudo apt-get update
sudo apt-get -y install build-essential zlib1g zlib1g-dev libssl-dev libpcre3 libpcre3-dev

pushd "${WORKING_DIR}"

  scream "Downloading upstream tarball"

  curl "${UPSTREAM_TARBALL}" \
    --silent \
    --output upstream.tgz

  UPSTREAM_SHA256=$(sha256sum upstream.tgz)
  UPSTREAM_SHA256=${UPSTREAM_SHA256:0:64}

  tar --extract \
    --file upstream.tgz

  pushd "nginx-${VERSION}"
    scream "Running NGINX's ./configure script"

    # static options
    # --with-cc-opt=-fPIE -pie
    # --with-ld-opt=-fPIE -pie -z now

    ./configure \
      --prefix=/ \
      --error-log-path=stderr \
      --with-http_ssl_module \
      --with-http_v2_module \
      --with-http_realip_module \
      --with-http_gunzip_module \
      --with-http_gzip_static_module \
      --with-http_auth_request_module \
      --with-http_random_index_module \
      --with-http_secure_link_module \
      --with-http_stub_status_module \
      --without-http_uwsgi_module \
      --without-http_scgi_module \
      --with-pcre \
      --with-pcre-jit \
      --with-debug \
      --with-cc-opt="-fPIC -pie" \
      --with-ld-opt="-fPIC -pie -z now" \
      --with-compat \
      --with-stream=dynamic \
      --with-http_sub_module

    scream "Running make and make install"

    make
    DESTDIR="${DEST_DIR}/nginx" make install
  popd
popd

pushd "${DEST_DIR}"
  rm -Rf ./nginx/html ./nginx/conf
  mkdir nginx/conf
  tar zcvf "${OUTPUT_DIR}/temp.tgz" .
popd

pushd "${OUTPUT_DIR}"

  SHA256=$(sha256sum temp.tgz)
  SHA256="${SHA256:0:64}"

  OUTPUT_TARBALL_NAME="nginx_${VERSION}_linux_x64_jammy_${SHA256:0:8}.tgz"

  scream "Building tarball ${OUTPUT_TARBALL_NAME}"

  mv temp.tgz "${OUTPUT_TARBALL_NAME}"
popd

echo "::set-output name=upstream-sha256::${UPSTREAM_SHA256}"
echo "::set-output name=tarball-name::${OUTPUT_TARBALL_NAME}"
echo "::set-output name=tarball-path::${OUTPUT_DIR}/${OUTPUT_TARBALL_NAME}"
echo "::set-output name=sha256::${SHA256}"
