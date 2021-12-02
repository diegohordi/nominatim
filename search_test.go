package nominatim_test

import (
	"context"
	"encoding/json"
	"github.com/diegohordi/nominatim"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"
)

func mustLoadValidSearchResults(t *testing.T) []byte {
	t.Helper()
	content, err := os.ReadFile("./test/testdata/valid_search_results.json")
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func mustLoadValidSearchResultsAsSlice(t *testing.T) []nominatim.Result {
	t.Helper()
	var results []nominatim.Result
	if err := json.Unmarshal(mustLoadValidSearchResults(t), &results); err != nil {
		t.Fatal(err)
	}
	return results
}

func Test_Search(t *testing.T) {
	type fields struct {
		baseURL string
		client  func() *http.Client
	}
	type args struct {
		ctx   func() (context.Context, context.CancelFunc)
		query func() nominatim.SearchQuery
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []nominatim.Result
		wantErr bool
	}{
		{
			name: "should fail due to context timeout",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							time.Sleep(10 * time.Second)
							return &http.Response{}
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.WithTimeout(context.TODO(), 1*time.Millisecond)
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					return *query
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "should fail due to client timeout",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Timeout: 1 * time.Millisecond,
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					return *query
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "should fail due to unknown body",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.WriteString("{}")
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					return *query
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "should retrieve results with free form query",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidSearchResults(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					return *query
				},
			},
			want:    mustLoadValidSearchResultsAsSlice(t),
			wantErr: false,
		},
		{
			name: "should retrieve results with a complete structured query",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidSearchResults(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.Street = "test"
					query.City = "test"
					query.County = "test"
					query.State = "test"
					query.Country = "test"
					query.PostalCode = "test"
					return *query
				},
			},
			want:    mustLoadValidSearchResultsAsSlice(t),
			wantErr: false,
		},
		{
			name: "should retrieve results with a valid limit",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidSearchResults(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					query.Limit = 10
					return *query
				},
			},
			want:    mustLoadValidSearchResultsAsSlice(t),
			wantErr: false,
		},
		{
			name: "should retrieve results with a limit < 0",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidSearchResults(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					query.Limit = -10
					return *query
				},
			},
			want:    mustLoadValidSearchResultsAsSlice(t),
			wantErr: false,
		},
		{
			name: "should retrieve results with a limit > 50",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidSearchResults(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					query.Limit = 100
					return *query
				},
			},
			want:    mustLoadValidSearchResultsAsSlice(t),
			wantErr: false,
		},
		{
			name: "should retrieve results without extra tags",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidSearchResults(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					query.ExtraTags = false
					return *query
				},
			},
			want:    mustLoadValidSearchResultsAsSlice(t),
			wantErr: false,
		},
		{
			name: "should retrieve results without name details",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidSearchResults(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					query.NameDetails = false
					return *query
				},
			},
			want:    mustLoadValidSearchResultsAsSlice(t),
			wantErr: false,
		},
		{
			name: "should retrieve results without address details",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidSearchResults(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					query.AddressDetails = false
					return *query
				},
			},
			want:    mustLoadValidSearchResultsAsSlice(t),
			wantErr: false,
		},
		{
			name: "should retrieve results with a list of accepted languages",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidSearchResults(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					query.AcceptLanguage = []string{"en", "pt"}
					return *query
				},
			},
			want:    mustLoadValidSearchResultsAsSlice(t),
			wantErr: false,
		},
		{
			name: "should retrieve results with a list of excluded places ID",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidSearchResults(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "test"
					query.ExcludedPlaces = []string{"123", "345"}
					return *query
				},
			},
			want:    mustLoadValidSearchResultsAsSlice(t),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := nominatim.NewClient(tt.fields.baseURL, tt.fields.client())
			ctx, cancelFn := tt.args.ctx()
			if cancelFn != nil {
				defer cancelFn()
			}
			got, err := d.Search(ctx, tt.args.query())
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Search() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Integration_Search(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests...")
	}
	type fields struct {
		baseURL string
		client  func() *http.Client
	}
	type args struct {
		ctx   context.Context
		query func() nominatim.SearchQuery
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should return results successfully",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Timeout: time.Second * 5,
					}
				},
			},
			args: args{
				ctx: context.TODO(),
				query: func() nominatim.SearchQuery {
					query := nominatim.NewSearchQuery()
					query.FreeFormQuery = "avenida da república, lisboa"
					return *query
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := nominatim.NewClient(tt.fields.baseURL, tt.fields.client())
			_, err := d.Search(tt.args.ctx, tt.args.query())
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
