#!/bin/bash

# this script allows to run dockbuild inside a container, without the user noticing that.

docker run -ti --rm --entrypoint /app/dockbuild.py -v /var/run/docker.sock:/var/run/docker.sock -v ${TARGET_DIR}/docker-${DOCKER_VERSION}:/usr/bin/docker mgrast/dockbuild $@
