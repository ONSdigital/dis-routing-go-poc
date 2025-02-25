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

To add new routes, the admin ap allows you to post JSON to `/routes` or `/redirects` to add new configurations to the
running router. The spec is contained in an OpenAPI spec in the `storage` package of this repo.

### POC specific code

Due to the POC nature, all code should be considered non-production. However, specific sctions of the code that are
very much POC specific are marked with comments.

```go
// POC-ONLY
…
// /POC-ONLY
```

Some examples include…

- Adding in debugging headers to requests
- Deliberate delays supplied via request headers

### Known / Potential issues

There is currently no validation on routes added by the POC admin api. This means invalid routes that don't conform to
the Go `http.ServeMux` routing pattern aren't picked up until the router component tries to apply them. As a result, the
router panics and the POC service terminates. This can be fixed in a production release by appying appropriate
validation as well as trapping the panic appropriately on reload
[similar to the deprecation middleware in the dp-api-router](https://github.com/ONSdigital/dp-api-router/blob/v2.25.0/deprecation/config.go#L86).

Also, the use of the standard library mux means
that [conflicting routes use its precedence rules](https://pkg.go.dev/net/http#hdr-Precedence-ServeMux).
For example, if a route or redirect uses a path of `/something/{id}` and another uses a path of `/{any}/blah` then
`/something/blah` will match both rules, but as neither rule is more specific than the other, so it is ambiguous and
hence the Go router will reject the adding of one of these rules. some other routers use a 'first match' disambiguation
rule so might be an alternative choice.  

### Dependencies

None. The POC is built using the Go standard library where possible.

### Configuration

None. Ports are hardcoded currently as it's only a POC.

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2025, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
