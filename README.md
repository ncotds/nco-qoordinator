# NC OMNIbus ObjectServer Query Coordinator #

HTTP-API to proxy requests to one or more "OMNIbus Object Server" instances and collects all results. 

"OMNIbus Object Server" - component of IBM Netcool stack, in-memory database to store alerts data


## Installation

You can run NCO-Qoordinator API server:
* [as docker container](docs/deploy-docker/README.md)
* [as systemd service](docs/deploy-systemd/README.md)

## Usage

#### NCO-Qoordinator API

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

## Versioning

We use [SemVer](http://semver.org/) for versioning.
For the versions available, see the [tags on this repository](https://github.com/ncotds/nco-qoordinator/tags). 


## TODO

- [ ] Try to switch from [freetds wrapper](https://github.com/minus5/gofreetds) to the Go-native driver
- [ ] Implement cli tool to run queries from terminal

## Developing

Prerequsites:

* [go 1.22+](https://go.dev/doc/install)
* [docker-ce, docker-compose](https://docs.docker.com/engine/install/)
* [pre-commit tool](https://pre-commit.com/#install)
* [freetds](https://www.freetds.org/index.html) (freetds-dev package)

Setup dev environment:

* clone repo and go to the project's root
* setup OMNIbus - TODO
* install tools and enable pre commit hooks:
  ```
  go mod download -x 
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