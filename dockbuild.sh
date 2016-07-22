#!/bin/bash

# this script allows to run dockbuild inside a container, without the user noticing that.



# This script will download a statically compiled docker binary, which can be easilt mounted into containers.
# The docker binary that comes with CoreOS is currently not statically compiled.

set -e


export TARGET_DIR="/media/ephemeral"
export DOCKER_VERSION=$(/usr/bin/docker --version | grep -o '[0-9]*\.[0-9]*\.[0-9]')

if [ ! -e ${TARGET_DIR}/docker-${DOCKER_VERSION} ] ; then
    if [[ $EUID -ne 0 ]]; then
        echo "This script must be run as root to be able to write to ${TARGET_DIR}" 1>&2
        exit 1
    fi
    set -x
    rm -f ${TARGET_DIR}/docker-${DOCKER_VERSION}_part
    curl --silent -o ${TARGET_DIR}/docker-${DOCKER_VERSION}_part --retry 10 https://get.docker.com/builds/Linux/x86_64/docker-${DOCKER_VERSION}
    chmod +x ${TARGET_DIR}/docker-${DOCKER_VERSION}_part
    mv ${TARGET_DIR}/docker-${DOCKER_VERSION}_part ${TARGET_DIR}/docker-${DOCKER_VERSION}
    set +x
    echo "Downloaded: ${TARGET_DIR}/docker-${DOCKER_VERSION}"
fi



docker run -ti --rm --entrypoint /app/dockbuild.py -v /var/run/docker.sock:/var/run/docker.sock -v ${TARGET_DIR}/docker-${DOCKER_VERSION}:/usr/bin/docker mgrast/dockbuild $@
