package nominatim

import (
	"net/url"
	"strings"
)

// ReverseQuery holds the parameters needed to perform the search.
type ReverseQuery struct {
	Latitude       string
	Longitude      string
	AddressDetails bool
	ExtraTags      bool
	NameDetails    bool
	AcceptLanguage []string
}

// NewReverseQuery creates a ReverseQuery with default values and the given options.
func NewReverseQuery(latitude, longitude string) *ReverseQuery {
	return &ReverseQuery{
		Latitude:       latitude,
		Longitude:      longitude,
		AcceptLanguage: []string{"en"},
		AddressDetails: true,
	}
}

// buildQueryString builds a query string accordingly with the given ReverseQuery.
func (q ReverseQuery) buildQueryString() string {
	queryStr := url.Values{}
	queryStr.Set(keyFormat, defaultFormat)
	queryStr.Set(keyLatitude, q.Latitude)
	queryStr.Set(keyLongitude, q.Longitude)
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
	return queryStr.Encode()
}
