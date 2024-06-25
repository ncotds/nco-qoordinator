### Running NCO-Qoordinator API as docker container...

...is the easiest way to run API-server:

* pull image:
  ```shell
  docker pull ghcr.io/ncotds/ncoq-qoordinator:latest
  ```
* run application:
  ```shell
  docker run \
    -e NCOQ_OMNI_CLUSTERS="AGG1:localhost:4100|localhost:4101" \
    -e NCOQ_OMNI_MAX_CONN="10" \
    -e NCOQ_OMNI_FAILBACK="true" \
    -e NCOQ_OMNI_FAILBACK_DELAY="300s" \
    --name=ncoq-api \
    --publish="4000:4000" \
    ncoq-qoordinator:latest
  ```

See available versions [here](https://github.com/ncotds/nco-qoordinator/pkgs/container/nco-qoordinator)

### Config

ENV variables to config api-server:
```shell
# if log_file is empty, sends log to STDOUT
NCOQ_LOG_FILE=""
NCOQ_LOG_LEVEL="ERROR"
#
# timeouts prevents 'Slowloris Attack'
#
# timeout is the maximum duration:
#   - for reading the entire request, including the body
#   - before timing out writes of the response
# a zero or negative value means there will be no timeout
NCOQ_HTTP_TIMEOUT="5s"
# idle_timeout is the maximum amount of time to wait for the next request
# when keep-alives are enabled
NCOQ_HTTP_IDLE_TIMEOUT="60s"
#
# use comma to split OMNI-clusters and | to split nodes in cluster
NCOQ_OMNI_CLUSTERS="AGG1:localhost:4100|localhost:4101"
#
# connection_label you can see in OMNIbus "catalog.connections" table (AppName field)
NCOQ_OMNI_CONN_LABEL="nco-qoordinator"
#
# max_connections can be established to each of clusters
NCOQ_OMNI_MAX_CONN="10"
#
# failover_policy
#
# random_fail_over sets clients fail over strategy to 'RandomFailOver'.
# it means that when current connection is loosed, client firstly tries to connect to any random address
# from seed list, except the current one. if those address fails, then client continue with next seed...
# and makes one attempt to each address from seed list until success or attempts to all addresses will fail
#random_fail_over: false
#
# fail_back sets clients failover policy to 'FailBack', overrides random_fail_over.
# it means that when current connection is loosed, client firstly tries to connect:
#   - to the first address from seed list if Delay exceeded (from the last reconnect try)
#   - to address from seed list, next to current one otherwise
# if those address fails, then client continue with next seed...
# and makes one attempt to each address from seed list until success or attempts to all addresses will fail
#fail_back: true
#fail_back_delay: 300s
NCOQ_OMNI_RAND_FAILOVER="false"
NCOQ_OMNI_FAILBACK="true"
NCOQ_OMNI_FAILBACK_DELAY="300s"
```
