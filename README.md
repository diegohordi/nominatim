# Nominatim Go Client

Go client for [Nominatim API](https://nominatim.org/).

## Usage

### Install

`go get github.com/diegohordi/nominatim`

### Creating a client

In order to use this client, you need to create your own http.Client, with your set of configurations, plus the base URL
where the Nominatim API is serving:

```
import "github.com/diegohordi/nominatim"
...
httpClient := &http.Client{Timeout: time.Second * 5}
apiURL := "http://localhost:8080"
client := nominatim.NewClient(apiURL, httpClient)
```

#### Timeouts

If you need a different timeout from the base client that you created, you can create a `context.WithTimeout` and pass
it as parameter to the endpoints handlers, as they are able to deal with context signalling too.

### /search

In order to user [Search API](https://nominatim.org/release-docs/latest/api/Search/) you need to create the query model
based on the structs:

```
type SearchStructuredQuery struct {
	Street     string
	City       string
	County     string
	State      string
	Country    string
	PostalCode string
}

type SearchQuery struct {
	SearchStructuredQuery
	FreeFormQuery  string
	AddressDetails bool
	ExtraTags      bool
	NameDetails    bool
	AcceptLanguage []string
	ExcludedPlaces []string
	Limit          int
}
```

Note that you need to choose between a search using a Free-Form query or use a Structured Query instead, as per API
documentation. Even though you pass both, Free-Form query will be prioritized. Also, note that if you pass a `limit` out 
of the valid range, the default (limit < 0) or maximum (limit > 50) limit will be sent. So, after planned the way you
use the Search API, you can do as follows:

```
query := nominatim.NewSearchQuery()
query.FreeFormQuery = "avenida da república, lisboa"
results, err := client.Search(ctx, *query)
```

### /reverse

To use [Reverse API](https://nominatim.org/release-docs/latest/api/Reverse/), also you need to create the query model 
base on the struct bellow:

```
type ReverseQuery struct {
	Latitude       string
	Longitude      string
	AddressDetails bool
	ExtraTags      bool
	NameDetails    bool
	AcceptLanguage []string
}
```

Both latitude and longitude are mandatory fields, and you must fill them with valid values, otherwise you'll 
receive an error. So, you can do as follows: 

```
query := nominatim.NewReverseQuery("38.6945252", "-9.3221278")
query.AddressDetails = false
result, err := client.Reverse(ctx, *query)
```

### /status

[Status API](https://nominatim.org/release-docs/latest/api/Status/) allows you to check the service status. To do that,
you don't need to fill any query model, just call the function as follows:

```
status, err := d.CheckStatus(ctx)
```

## Tests

The coverage so far is greater than 95%, covering also failure scenarios. Also, as the handlers are dealing with context
timeout, there are no race conditions detected in the -race tests.

You can run the short test and the race condition test from Makefile, as below:

### Short
`make test_short`

### Race
`make test_race`

### Integration

There are integration tests available too. In order to run them properly, you'll need to, first, start a local Nominatim
server. There's one available from `./deployments/docker-compose.yml` file and you can start it also from Makefile, but
be advised that it takes too long time to start.

To start the Nominatim server:

`make start_dev_env`

After start the server and make sure that the Nominatim API is serving correctly, you can run the integration tests,
removing -short flag as:

`go test -count=1 ./...`

## TODO

- [ ] Support formats GEOJSON and GEOCODEJSON
- [ ] Implement [Address Lookup](https://nominatim.org/release-docs/latest/api/Lookup/) endpoint
- [ ] Implement [Details](https://nominatim.org/release-docs/latest/api/Details/) endpoint
- [ ] Automate integration tests
- [ ] Benchmark tests
- [ ] ...
