# dis-routing-go-poc

Proof of Concept for dynamically loadable routing and redirects in Go

## Getting started

The POC consists of three main parts:

- The router itself (mainly in the `routing` subdirectory, port 30000)
- A stubbed upstream service to record requests routed by the router (`upstream`, port 30001)
- A storage mock with its own simple admin API (`storage`, port 30002)

```shell
go run main.go
```

You can then call routes on the router using a browser or your http client of choice.
Eg. http://localhost:30000/some/route

### POC specific code

Due to the POC nature,  all code should be considered non-production. However, specific sctions of the code that are 
very much POC specific are marked with comments.  
```go
// POC-ONLY
…
// /POC-ONLY
```

Some examples include…

- Adding in debugging headers to requests
- Deliberate delays supplied via request headers

### Dependencies

None. The POC is built using the Go standard library where possible.

### Configuration

None. Ports are hardcoded currently as it's only a POC but this may change as the POC develops

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2025, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
