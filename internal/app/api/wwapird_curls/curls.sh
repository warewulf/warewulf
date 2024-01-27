#! /usr/bin/env bash

# This file is a scratchpad for curling wwapird.

# version
curl http://localhost:9871/version

# secure version
curl --cacert /usr/local/etc/warewulf/keys/cacert.pem \
    --key /usr/local/etc/warewulf/keys/client.key \
    --cert /usr/local/etc/warewulf/keys/client.pem \
    https://localhost:9871/version

# container list all
curl http://localhost:9871/v1/container

# container import
curl -d '{"source": "docker://ghcr.io/warewulf/warewulf-rockylinux:8", "name": "rocky-8", "update": true, "default": true}' -H "Content-Type: application/json" -X POST http://localhost:9871/v1/container

# container delete
curl -X DELETE http://localhost:9871/v1/container?containerNames=rocky-8

# container build
curl -d '{"containerNames": ["rocky-8"], "force": true}' -H "Content-Type: application/json" -X POST http://localhost:9871/v1/containerbuild

# node list all
curl http://localhost:9871/v1/node

# node list one
curl http://localhost:9871/v1/node?nodeNames=testnode1 # this works! case sensitive

# This is a list of testnode[1-2] with URL escapes.
curl http://localhost:9871/v1/node?nodeNames=testnode%5B1-2%5D

# node add single discoverable node
curl -d '{"nodeNames": ["testApiNode0"], "discoverable": true}' -H "Content-Type: application/json" -X POST http://localhost:9871/v1/node

curl -d '{"nodeNames": ["testApiNode1"], "discoverable": true}' -H "Content-Type: application/json" -X POST http://localhost:9871/v1/node


# list the node we just added
curl http://localhost:9871/v1/node?nodeNames=testApiNode0

# This gets me a little farther, but still no param data:
curl -d '{"nodeNames": ["testApiNode0"], "ipmiIpAddr": "10.0.8.220", "updateMask": "ipmiIpAddr,nodeNames"}' -H "Content-Type: application/json" -X PATCH http://localhost:9871/v1/node

# Node set with post:
curl -d '{"nodeNames": ["testApiNode0"], "ipmiIpaddr": "6.7.8.9"}' -H "Content-Type: application/json" -X POST http://localhost:9871/v1/nodeset

# node status
curl http://localhost:9871/v1/nodestatus

curl http://localhost:9871/v1/nodestatus?nodeNames=testApiNode0

# node delete single node
curl -X DELETE http://localhost:9871/v1/node?nodeNames=testApiNode0
curl -X DELETE http://localhost:9871/v1/node?nodeNames=testApiNode1
