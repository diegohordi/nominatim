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

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func mustLoadValidStatus(t *testing.T) []byte {
	t.Helper()
	content, err := os.ReadFile("./test/testdata/valid_status.json")
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func mustLoadValidStatusAsStruct(t *testing.T) nominatim.Status {
	t.Helper()
	status := &nominatim.Status{}
	if err := json.Unmarshal(mustLoadValidStatus(t), &status); err != nil {
		t.Fatal(err)
	}
	return *status
}

func Test_CheckStatus(t *testing.T) {
	type fields struct {
		baseURL string
		client  func() *http.Client
	}
	type args struct {
		ctx func() (context.Context, context.CancelFunc)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    nominatim.Status
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
			},
			want:    nominatim.Status{},
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
			},
			want:    nominatim.Status{},
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
							resp.Body.WriteString("[]")
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
			},
			want:    nominatim.Status{},
			wantErr: true,
		},
		{
			name: "should retrieve a valid status",
			fields: fields{
				baseURL: "http://localhost:8080",
				client: func() *http.Client {
					return &http.Client{
						Transport: RoundTripFunc(func(req *http.Request) *http.Response {
							resp := httptest.NewRecorder()
							resp.Body.Write(mustLoadValidStatus(t))
							return resp.Result()
						}),
					}
				},
			},
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.TODO(), nil
				},
			},
			want:    mustLoadValidStatusAsStruct(t),
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
			got, err := d.CheckStatus(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Integration_CheckStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests...")
	}
	type fields struct {
		baseURL string
		client  func() *http.Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should return status successfully",
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
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := nominatim.NewClient(tt.fields.baseURL, tt.fields.client())
			_, err := d.CheckStatus(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
