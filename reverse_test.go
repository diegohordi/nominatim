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

func mustLoadInvalidReverseResult(t *testing.T) []byte {
	t.Helper()
	content, err := os.ReadFile("./test/testdata/invalid_reverse_result.json")
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func mustLoadValidReverseResult(t *testing.T) []byte {
	t.Helper()
	content, err := os.ReadFile("./test/testdata/valid_reverse_result.json")
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func mustLoadValidReverseResultAsStruct(t *testing.T) nominatim.Result {
	t.Helper()
	result := &nominatim.Result{}
	if err := json.Unmarshal(mustLoadValidReverseResult(t), &result); err != nil {
		t.Fatal(err)
	}
	return *result
}

func Test_Reverse(t *testing.T) {
	type fields struct {
		baseURL string
		client  func() *http.Client
	}
	type args struct {
		ctx   func() (context.Context, context.CancelFunc)
		query func() nominatim.ReverseQuery
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    nominatim.Result
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
				query: func() nominatim.ReverseQuery {
					query := nominatim.NewReverseQuery("38.6945252", "-9.3221278")
					return *query
				},
			},
			want:    nominatim.Result{},
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
				query: func() nominatim.ReverseQuery {
					query := nominatim.NewReverseQuery("38.6945252", "-9.3221278")
					return *query
				},
			},
			want:    nominatim.Result{},
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
							resp.Body.WriteString("[{}]")
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.ReverseQuery {
					query := nominatim.NewReverseQuery("38.6945252", "-9.3221278")
					return *query
				},
			},
			want:    nominatim.Result{},
			wantErr: true,
		},
		{
			name: "should fail due to invalid latitude and longitude",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadInvalidReverseResult(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.ReverseQuery {
					query := nominatim.NewReverseQuery("test", "testing")
					return *query
				},
			},
			want:    nominatim.Result{},
			wantErr: true,
		},
		{
			name: "should retrieve results without extra tags",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidReverseResult(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.ReverseQuery {
					query := nominatim.NewReverseQuery("", "")
					query.ExtraTags = false
					return *query
				},
			},
			want:    mustLoadValidReverseResultAsStruct(t),
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
							resp.Body.Write(mustLoadValidReverseResult(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.ReverseQuery {
					query := nominatim.NewReverseQuery("38.6945252", "-9.3221278")
					query.NameDetails = true
					return *query
				},
			},
			want:    mustLoadValidReverseResultAsStruct(t),
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
							resp.Body.Write(mustLoadValidReverseResult(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.ReverseQuery {
					query := nominatim.NewReverseQuery("38.6945252", "-9.3221278")
					query.AddressDetails = false
					return *query
				},
			},
			want:    mustLoadValidReverseResultAsStruct(t),
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
							resp.Body.Write(mustLoadValidReverseResult(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
				query: func() nominatim.ReverseQuery {
					query := nominatim.NewReverseQuery("38.6945252", "-9.3221278")
					query.AcceptLanguage = []string{"en", "pt"}
					return *query
				},
			},
			want:    mustLoadValidReverseResultAsStruct(t),
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
			got, err := d.Reverse(ctx, tt.args.query())
			if (err != nil) != tt.wantErr {
				t.Errorf("Reverse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reverse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Integration_Reverse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests...")
	}
	type fields struct {
		baseURL string
		client  func() *http.Client
	}
	type args struct {
		ctx   context.Context
		query func() nominatim.ReverseQuery
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should return result successfully",
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
				query: func() nominatim.ReverseQuery {
					query := nominatim.NewReverseQuery("38.6945252", "-9.3221278")
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
			_, err := d.Reverse(tt.args.ctx, tt.args.query())
			if (err != nil) != tt.wantErr {
				t.Errorf("Reverse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
