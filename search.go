package nominatim

import (
	"net/url"
	"strconv"
	"strings"
)

// SearchStructuredQuery holds parameters used to perform a structured query.
type SearchStructuredQuery struct {
	Street     string
	City       string
	County     string
	State      string
	Country    string
	PostalCode string
}

// SearchQuery holds the parameters needed to perform the search.
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

// NewSearchQuery creates a SearchQuery with default values and the given options.
func NewSearchQuery() *SearchQuery {
	return &SearchQuery{
		Limit:          10,
		AcceptLanguage: []string{"en"},
		AddressDetails: true,
	}
}

// buildQueryString builds a query string accordingly with the given SearchQuery.
func (q SearchQuery) buildQueryString() string {
	queryStr := url.Values{}
	queryStr.Set(keyFormat, defaultFormat)
	if q.FreeFormQuery != "" {
		queryStr.Set(keyFreeFormQuery, q.FreeFormQuery)
	}
	if q.FreeFormQuery == "" && q.Street != "" {
		queryStr.Set(keyStreet, q.Street)
	}
	if q.FreeFormQuery == "" && q.City != "" {
		queryStr.Set(keyCity, q.City)
	}
	if q.FreeFormQuery == "" && q.County != "" {
		queryStr.Set(keyCounty, q.County)
	}
	if q.FreeFormQuery == "" && q.State != "" {
		queryStr.Set(keyState, q.State)
	}
	if q.FreeFormQuery == "" && q.Country != "" {
		queryStr.Set(keyCountry, q.Country)
	}
	if q.FreeFormQuery == "" && q.PostalCode != "" {
		queryStr.Set(keyPostalCode, q.PostalCode)
	}
	queryStr.Set(keyAddressDetails, "1")
	if !q.AddressDetails {
		queryStr.Set(keyAddressDetails, "0")
	}
	queryStr.Set(keyExtraTags, "1")
	if !q.ExtraTags {
		queryStr.Set(keyExtraTags, "0")
	}
	queryStr.Set(keyNameDetails, "1")
	if !q.NameDetails {
		queryStr.Set(keyNameDetails, "0")
	}
	if len(q.AcceptLanguage) > 0 {
		queryStr.Set(keyAcceptLanguage, strings.Join(q.AcceptLanguage, ","))
	}
	if len(q.ExcludedPlaces) > 0 {
		queryStr.Set(keyExcludePlaces, strings.Join(q.ExcludedPlaces, ","))
	}
	if q.Limit != 0 {
		limit := q.Limit
		if limit < 0 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}
		queryStr.Set(keyLimit, strconv.Itoa(limit))
	}
	return queryStr.Encode()
}
