# NC OMNIbus ObjectServer Query Coordinator #

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=ncotds_nco-qoordinator&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=ncotds_nco-qoordinator)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=ncotds_nco-qoordinator&metric=coverage)](https://sonarcloud.io/summary/new_code?id=ncotds_nco-qoordinator)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=ncotds_nco-qoordinator&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=ncotds_nco-qoordinator)
[![Build](https://github.com/ncotds/nco-qoordinator/actions/workflows/build-release-assets.yml/badge.svg)](https://github.com/ncotds/nco-qoordinator/actions/workflows/build-release-assets.yml)

> *"OMNIbus Object Server" - component of IBM Netcool stack, in-memory database to store alerts data*

HTTP-API server to proxy requests to one or more OMNIbus ObjectServer instances and collects all results. 

The application was created to address a specific need:
in environments with multiple shards of the OMNIbus database (several HA clusters where alerts are distributed), 
third-party applications require a solution to seamlessly access and update alert data across all shards.

Additionally, this application proves beneficial even in scenarios with a single HA cluster. 
It supports multiple third-party applications by allowing them to interact with alerts consistently, 
eliminating the need to locate a suitable OMNIbus database driver and manage failover scenarios.

## Installation

You can run NCO-Qoordinator API server:
* [as docker container](docs/deploy-docker/README.md)
* [as systemd service](docs/deploy-systemd/README.md)

## Usage

### NCO-Qoordinator API

#### Request examples

* run SQL-statement on all clusters
  ```shell
  curl -X POST \
    -H 'Content-Type: application/json' \
    -H 'X-Request-Id: 848dc8a1-0e8e-454c-b1de-d0d9dd6df439' \
    -u root:strong \
    --data '{"sql": "select top 3 Node, Severity, Summary from status where Type=1"}' \
  http://localhost:4000/rawSQL | jq
  ```
  ```json
  [
    {
      "clusterName": "AGG2",
      "rows": [
        {
          "Node": "dlbjnwg.net",
          "Severity": 5,
          "Summary": "Eum asperiores vero ut harum architecto."
        },
        {
          "Node": "xqunkrh.edu",
          "Severity": 3,
          "Summary": "Dicta aspernatur aut iure et voluptatum."
        },
        {
          "Node": "prcxdcd.org",
          "Severity": 5,
          "Summary": "Voluptatem et et vel omnis magni."
        }
      ],
      "affectedRows": 3
    },
    {
      "clusterName": "AGG1",
      "rows": [
        {
          "Node": "otqrdsd.top",
          "Severity": 4,
          "Summary": "Eum earum debitis aut in voluptatem."
        },
        {
          "Node": "njcjymg.edu",
          "Severity": 3,
          "Summary": "Iste fugiat et rerum perspiciatis voluptas."
        },
        {
          "Summary": "Eum asperiores vero ut harum architecto.",
          "Node": "dlbjnwg.net",
          "Severity": 5
        }
      ],
      "affectedRows": 3
    }
  ]
  ```
* run SQL-statement on given cluster only
  ```shell
  curl -X POST \
    -H 'Content-Type: application/json' \
    -H 'X-Request-Id: b16d95e3-28ef-493f-89ab-02c135ceb365' \
    -u root:strong \
    --data "{\"sql\": \"update status set Acknowledged = 1 where Type = 1 and Node = 'njcjymg.edu'\",\"clusters\":[\"AGG1\"]}" \
  http://localhost:4000/rawSQL | jq
  ```
  ```json
  [
    {
      "clusterName": "AGG1",
      "rows": [],
      "affectedRows": 1
    }
  ]
  ```
* get configured cluster names
  ```shell
  curl -X GET \
  -H 'Content-Type: application/json' \
  -H 'X-Request-Id: cf67a325-065c-4f4c-9600-64245b82f615' \
  -u root:strong \
  http://localhost:4000/clusterNames | jq
  ```
  ```json
  [
    "AGG1",
    "AGG2"
  ]
  ```
  
#### Connection pool

The API server opens connections to OMNIbus as needed,
credentials from the request authorization header are used for login.

After the request is completed, the connection is returned to the pool. 
The maximum number of simultaneously open connections is set in the [configuration](config/example.yml).

A separate pool of connections is created for each OMNIbus cluster.

If the maximum number of connections has been reached, but to complete the request 
it is necessary to open a new one for the current user, the `ncoq-api` server will close 
the oldest unused connection in the pool. 

If all connections are used, an error will be returned

#### Failover, failback and load balancing

The [configuration](config/example.yml) allows you to specify several instances for each OMNIbus cluster, 
in case of failure, the server will try to connect to another instance.

There are two strategies for selecting an OMNIbus instance to reconnect:
1) Random failover (`random_fail_over: true`) - the server will select any of the instances 
   except the current one. If it is unavailable, it will try all options one by one 
   until successful connection. Will be useful for OMNIbus Display level
2) Failback (`fail_back: true, fail_back_delay: 300s`) - the server tries to connect to 
   the next OMNIbus instance from the list, and if the `fail_back_delay` interval has expired, 
   it tries turn back to the main one (the first one in the list). 
   Useful for Aggregation level

## Versioning

We use [SemVer](http://semver.org/) for versioning.
For the versions available, see the [tags on this repository](https://github.com/ncotds/nco-qoordinator/tags). 


## TODO

- [x] ~~Try to switch from [freetds wrapper](https://github.com/minus5/gofreetds) to the Go-native driver~~

## Developing

Prerequsites:

* [go 1.22+](https://go.dev/doc/install)
* [docker-ce, docker-compose](https://docs.docker.com/engine/install/)
* [pre-commit tool](https://pre-commit.com/#install)

Setup dev environment:

* clone repo and go to the project's root
* setup OMNIbus
  (if you prefer docker, 
   see the [repo with Dockerfiles for Netcool](https://github.com/juliusloman/docker-omnibus)
   and example [docker-compose file](tests/docker-compose-omni.yml))
* install tools and enable pre commit hooks:
  ```
  make setup-tools 
  pre-commit install
  ```
* run tests:
  ```
  make lint test
  ```
* setup and run:
  ```
  cp config/example.yml local.yml  # set actual values
  make run-ncoq-api
  ```