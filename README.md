# Automated Approver

The Automated Approver is an agent that can operate as a standalone service (exposing a RESTful API to 3rd party applications), or as a library integrated into an application.

More details on how to use the automated approver can be found in `/docs`.

## How to build
```shell
> make build
```

## Running Tests
```shell
> make test
```
will run all unit and restAPI tests.

```shell
> make unittest
```
runs just unit test.
```shell
> make apitest
```
will run the restAPI test.

### End to end test (e2e)
In order to run the e2e test, the APIKEY and BASE64PKEY (the base64 of the private.pem file) for a Qeedo account to test
against are needed.  Ensure both the APIKEY and BASE64PKEY are set in the environment before running the e2e test.
```shell
> make teste2e
```
