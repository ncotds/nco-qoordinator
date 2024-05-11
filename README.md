# NC OMNIbus ObjectServer Query Coordinator #

HTTP-API to proxy requests to one or more "OMNIbus Object Server" instances and collects all results. 

"OMNIbus Object Server" - component of IBM Netcool stack, in-memory database to store alerts data


## Installation

TODO

## Usage


TODO


## Versioning

We use [SemVer](http://semver.org/) for versioning.
For the versions available, see the [tags on this repository](#). 


## TODO

TODO


## Developing

Prerequsites:

* [go 1.22+](https://go.dev/doc/install)
* [docker-ce, docker-compose](https://docs.docker.com/engine/install/)
* [pre-commit tool](https://pre-commit.com/#install)

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
* setup and run TODO
