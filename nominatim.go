package nominatim

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultFormat = "jsonv2"
)

const (
	endpointSearch  = "search"
	endpointReverse = "reverse"
	endpointStatus  = "status"
)

const (
	StatusNoDatabase       = 700
	StatusModuleFailed     = 701
	StatusModuleCallFailed = 702
	StatusQueryFailed      = 703
	StatusNoValue          = 704
)

const (
	keyAddressDetails = "addressdetails"
	keyExtraTags      = "extratags"
	keyNameDetails    = "namedetails"
	keyAcceptLanguage = "accept-language"
	keyExcludePlaces  = "exclude_place_ids"
	keyFreeFormQuery  = "q"
	keyStreet         = "street"
	keyCity           = "city"
	keyCounty         = "county"
	keyState          = "state"
	keyCountry        = "country"
	keyPostalCode     = "postalcode"
	keyLimit          = "limit"
	keyLatitude       = "lat"
	keyLongitude      = "lon"
	keyFormat         = "format"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// Address holds address information from a result.
type Address struct {
	City           string `json:"city"`
	CityDistrict   string `json:"city_district"`
	Construction   string `json:"construction"`
	Continent      string `json:"continent"`
	Country        string `json:"country"`
	CountryCode    string `json:"country_code"`
	HouseNumber    string `json:"house_number"`
	Neighbourhood  string `json:"neighbourhood"`
	Postcode       string `json:"postcode"`
	PublicBuilding string `json:"public_building"`
	State          string `json:"state"`
	Suburb         string `json:"suburb"`
}

// Result holds information from a specific location.
type Result struct {
	PlaceId     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmId       int      `json:"osm_id"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	PlaceRank   int      `json:"place_rank"`
	Category    string   `json:"category"`
	Type        string   `json:"type"`
	Importance  float64  `json:"importance"`
	AddressType string   `json:"addresstype"`
	DisplayName string   `json:"display_name"`
	Name        string   `json:"name"`
	Address     Address  `json:"address"`
	BoundingBox []string `json:"bounding_box"`
}

// Status holds information from Nomination API server.
type Status struct {
	Status          int       `json:"status"`
	Message         string    `json:"message"`
	DataUpdated     time.Time `json:"data_updated"`
	SoftwareVersion string    `json:"software_version"`
	DatabaseVersion string    `json:"database_version"`
}

type SearchHandler interface {

	// Search looks up a location from a textual description or address.
	Search(ctx context.Context, query SearchQuery) ([]Result, error)
}

type ReverseHandler interface {

	// Reverse generates an address from a latitude and longitude.
	Reverse(ctx context.Context, query ReverseQuery) (Result, error)
}

type StatusHandler interface {

	// CheckStatus checks if Nominatim service and database is running.
	CheckStatus(ctx context.Context) (Status, error)
}

type Client interface {
	SearchHandler
	ReverseHandler
	StatusHandler
}

type defaultClient struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string, client *http.Client) Client {
	return &defaultClient{baseURL: baseURL, client: client}
}

func (d defaultClient) Search(ctx context.Context, query SearchQuery) ([]Result, error) {
	resultsChan := make(chan []Result, 1)
	errChan := make(chan error, 1)
	endpoint := fmt.Sprintf("%s/%s?%s", d.baseURL, endpointSearch, query.buildQueryString())

	go func() {
		resp, err := d.client.Get(endpoint)
		if err != nil {
			errChan <- err
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		results := make([]Result, 0)
		if err = json.NewDecoder(resp.Body).Decode(&results); err != nil {
			errChan <- err
		}
		resultsChan <- results
	}()

	select {
	case results := <-resultsChan:
		return results, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (d defaultClient) Reverse(ctx context.Context, query ReverseQuery) (Result, error) {
	resultChan := make(chan Result, 1)
	errChan := make(chan error, 1)
	endpoint := fmt.Sprintf("%s/%s?%s", d.baseURL, endpointReverse, query.buildQueryString())

	go func() {
		resp, err := d.client.Get(endpoint)
		if err != nil {
			errChan <- err
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		result := &struct {
			Result
			Error Error `json:"error"`
		}{}
		if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
			errChan <- err
		}
		if result.Error.Code > 0 {
			errChan <- result.Error
		}
		resultChan <- result.Result
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return Result{}, err
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}

func (d defaultClient) CheckStatus(ctx context.Context) (Status, error) {
	statusChan := make(chan Status, 1)
	errChan := make(chan error, 1)
	endpoint := fmt.Sprintf("%s/%s?format=json", d.baseURL, endpointStatus)

	go func() {
		resp, err := d.client.Get(endpoint)
		if err != nil {
			errChan <- err
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		status := &Status{}
		if err = json.NewDecoder(resp.Body).Decode(status); err != nil {
			errChan <- err
		}
		statusChan <- *status
	}()

	select {
	case result := <-statusChan:
		return result, nil
	case err := <-errChan:
		return Status{}, err
	case <-ctx.Done():
		return Status{}, ctx.Err()
	}
}
