# DISCONTINUATION OF PROJECT #
This project will no longer be maintained by Intel.
Intel has ceased development and contributions including, but not limited to, maintenance, bug fixes, new releases, or updates, to this project.
Intel no longer accepts patches to this project.
# Utilities

![coverage|99.3%](https://img.shields.io/static/v1?label=coverage&message=99.3%&color=brightgreen&style=flat-square) ![license|BSD-3-Clause](https://img.shields.io/static/v1?label=license&message=BSD-3-Clause&color=blue&style=flat-square)

## Table of Contents

<!-- toc -->

- [Description](#description)
- [Documentation](#documentation)
- [Requirements](#requirements)
- [Setup](#setup)
- [Testing](#testing)
  * [Linting](#linting)
- [Configuration](#configuration)
- [Road Map & Discussion](#road-map--discussion)
  * [REST API Payload Discussion](#rest-api-payload-discussion)

<!-- tocstop -->

## Description

This package contains many functions and Go structs that are used by the [Automated Checkout reference design](https://github.com/intel-iot-devkit/automated-checkout). 

## Documentation

Each function in this package is documented. To see documentation, run the command `go doc -all`.

This package is mostly centered around building & writing instances of the `HTTPResponse` struct, which contains a payload like this:

```go
type HTTPResponse struct {
	Content     interface{} `json:"content"`
	ContentType string      `json:"contentType"`
	StatusCode  int         `json:"statusCode"`
	Error       bool        `json:"error"`
}
```

In JSON:

```json
{
	"content": "{}",
	"contentType": "string|json",
	"statusCode": 200,
	"error": false
}
```

Philosophically, this payload format has some advantages and disadvantages, which are discussed below in the [REST API Payload Discussion](#rest-api-payload-discussion) section.

This package is not restricted to exporting only HTTP response-like functions and structs. It can hold any code that is intended to be used with Automated Checkout and EdgeX.

## Requirements

This package has no dependencies or requirements aside from the essentials for Go development:

- Git version 2.22+ (or latest supported version)
- Go v1.12+

## Setup

This package is intended to be used as part of a `go.mod` file with a reference to a locally cloned version of this repository - here's a modified example of another repository's `go.mod` file:

```
module as-vending

go 1.12

require (
	github.com/edgexfoundry/app-functions-sdk-go v1.0.0
	github.com/edgexfoundry/go-mod-core-contracts v0.1.14
	github.com/intel-iot-devkit/automated-checkout-utilities v1.0.0
)

```

This step should be all that is required in order for your project to work with `automated-checkout-utilities`.

## Testing

To test, run:

```bash
make test
```

For an HTML readout of test coverage, run:

```bash
make testHTML
```

### Linting

The `Makefile` in this repository leverages [`golangci-lint`](https://github.com/golangci/golangci-lint) for in-depth linting, configured in `.golangci.yml`:

```bash
make lint
```

## Configuration

There is no configuration needed for this package.

### REST API Payload Discussion

REST API responses can be consistent when done right, for example, HTTP responses should have an accurate `Content-Type` header, such as `text/plain` or `application/json`. Additionally, response status codes such as 404, 200, and 500 help provide more useful information.

In the industry and in reality, REST API payload schemas are actually quite varied. For example, some companies always respond `200 OK` even for errors or non-OK responses, and then have a consistent schema in their response body that indicates the actual information. Other companies use variations on the theme.

One situation that demonstrates the flaws of using _only_ the HTTP status code is a `404` response. Does a `404` from the server mean that the endpoint does not exist, or that it was the correct endpoint, but the query yielded no results? This is one of the few but very frequently encountered cases of ambiguity that requires a bit more agreement of standardization across software projects.

In this particular project, the format of

```json
{
	"content": "{}",
	"contentType": "string|json",
	"statusCode": 200,
	"error": false
}
```

has been chosen because it is the bare minimum of what's needed for each of the microservices to interact with each other. This format solves the above scenario of an ambiguous 404 by allowing `error: true` when the endpoint is not found, and `error: false` when the endpoint is found but the query yielded no results. Additionally, all payloads are of the format `application/json`, and the `contentType` field dictates whether the `content` is a string or JSON response to be marshaled in Go.

There is some redundancy, but the consistency allows for smoother interactions between microservices.

**Disclaimer: The authors of the Automated Checkout reference design are not suggesting that all REST API schemas should follow this format.** Please use REST API interaction schemas that meet your own use cases best.
