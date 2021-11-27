# Warewulf daemon http interface

## runtime

### image retrieval
http GET request directed to
`http://<masterIP>:9873/overlay-runtime/`
will return the runtime overlay for the requesting IP, given certain conditions.
1. the request stems from Port `987`, unless the `secure` flag in the config is set to `false`
2. the IP is known to the master
3. a runtime overlay was built for the requesting node, or the `auto build` is set to `true` in the config.

## status

### general status information
Node status information is returned for GET requests to
`http://<masterIP>:9873/status/`

The information is - so far - limited to **node name**, **cluster name** and **last seen**.
* **last seen** is the time in seconds since the last checking for runtime overlay retrieval of the node.
* Limted to the period since the start/restart of the warewulf daemon.

The amount of information can be restricted with the `limit=<number of nodes>` parameter.

Example:
```shell
curl "http://localhost:9873/status/?limit=3  | w3m -T text/html -dump"
```
will report the three nodes with longest time since last chech-in.
```
NODE NAME              CLUSTER         last seen (s)
loki                                   34565
test_13                test            359
test_16                test            359
```

#### test data
For testing purposes, a number of test nodes can be added to the warewulf daemon data set.
```shell
curl  "http://localhost:9873/status/?test=40&limit=13  | w3m -T text/html -dump"
```
This will create 40 test nodes `test_0..39` in the cluster `test`.
But, only 13 of those will be shown.
A subsequent request with a lower number of nodes will retain the earlier instantiated nodes.
This can be used to mimic nodes that fail to check-in.
, the `limit` parameter takes no affect.
