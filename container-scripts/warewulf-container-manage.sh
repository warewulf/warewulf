#!/bin/bash
# quick script to manage the container
CONTAINER_NAME=warewulf

create_container() {
podman create \
    --name ${CONTAINER_NAME} \
    --tls-verify=false \
    --network host \
    ${IMAGE}
}

run_container() {
podman run \
    --name ${CONTAINER_NAME} \
    --rm -ti \
    --network host \
    --entrypoint bash \
    ${IMAGE}
}

if [ -z "$1" ]; then
echo "
First ARG is mandatory:
$0 [create|start|stop|rm|rmcache|run|bash|logs|install|uninstall]

CONTAINER_NAME: '${CONTAINER_NAME}'

DEPLOYMENT:
create
    Pull the image and create the container automatically

install
    install needed files on the host to manage '${CONTAINER_NAME}' container
    (in /usr/local/bin and /etc)

start
    Start the container '${CONTAINER_NAME}'

REMOVAL:
uninstall
    uninstall all needed files on the host to manage '${CONTAINER_NAME}' container

stop
    stop the container '${CONTAINER_NAME}'

rm
    delete the container '${CONTAINER_NAME}'

rmcache
    remove the container image in cache ${IMAGE}

DEBUG:
run
    podman run container '${CONTAINER_NAME}'

bash 
    go with /bin/bash command inside '${CONTAINER_NAME}'

logs
    see log of container '${CONTAINER_NAME}'

 "
 exit 1
fi

###########
# MAIN
###########
set -euxo pipefail

case $1 in
    start)
	podman start ${CONTAINER_NAME}
	podman ps | grep ${CONTAINER_NAME}
    ;;
    stop)
	podman stop ${CONTAINER_NAME}
	podman ps | grep ${CONTAINER_NAME}
    ;;
    rm)
    set +e
    podman stop ${CONTAINER_NAME}
    podman rm ${CONTAINER_NAME}
    ;;
    create)
    create_container
    ;;
    run)
    run_container
    ;;
    rmcache)
    podman rmi ${IMAGE}
    ;;
    logs)
    podman logs ${CONTAINER_NAME}
    ;;
    bash)
    set +e
    podman exec -ti ${CONTAINER_NAME} $@
    ;;
    install)
    podman run --env IMAGE=${IMAGE} --rm --privileged -v /:/host ${IMAGE} /bin/bash /container/label-install
    ;;
    uninstall)
    podman run --env IMAGE=${IMAGE} --rm --privileged -v /:/host ${IMAGE} /bin/bash /container/label-uninstall
    ;;
esac
