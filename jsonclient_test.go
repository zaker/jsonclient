package jsonclient

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

type ReqTestArgs[T any] struct {
	path string

	queryParameters url.Values

	responseType T
}

type TestRow[T any] struct {
	args    ReqTestArgs[T]
	want    T
	wantErr bool
}

var _ Requester = MockRequester{}
var _ context.Context = MockContext{}

type MockRequester struct {
}

// Do implements Requester.
func (MockRequester) Do(req *http.Request) (*http.Response, error) {

	if req.URL.Path == "" {

		return &http.Response{Body: io.NopCloser(strings.NewReader("{}")), StatusCode: 200}, nil
	}

	if req.URL.Path == "/foo" {

		return &http.Response{Body: io.NopCloser(strings.NewReader(`{"foo":"bar"}`)), StatusCode: 200}, nil
	}

	if req.URL.Path == "/slow" {
		time.Sleep(1000)
		return &http.Response{Body: io.NopCloser(strings.NewReader(`{"foo":"bar"}`)), StatusCode: 200}, nil
	}
	return &http.Response{Body: io.NopCloser(strings.NewReader("")), StatusCode: 500}, nil
}

type MockContext struct {
}

// Deadline implements context.Context.
func (MockContext) Deadline() (deadline time.Time, ok bool) {
	panic("unimplemented")
}

// Done implements context.Context.
func (MockContext) Done() <-chan struct{} {
	panic("unimplemented")
}

// Err implements context.Context.
func (MockContext) Err() error {
	panic("unimplemented")
}

// Value implements context.Context.
func (MockContext) Value(key any) any {
	panic("unimplemented")
}

func TestMakeHTTPRequest(t *testing.T) {

	tests := make(map[string]TestRow[map[string]any], 0)

	rh, err := NewClient[map[string]any]("http://example.com",
		WithClient[map[string]any](MockRequester{}),
		WithContext[map[string]any](MockContext{}),
		WithHeaders[map[string]any](map[string]string{"X-TEST": "FOO"}))

	if err != nil {
		t.Errorf("setting up test client: %v", err)
		return
	}

	tests["first"] = TestRow[map[string]any]{ReqTestArgs[map[string]any]{"", nil, nil}, map[string]any{}, false}
	tests["second"] = TestRow[map[string]any]{ReqTestArgs[map[string]any]{"/foo", nil, nil}, map[string]any{"foo": "bar"}, false}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := rh.GET(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeHTTPRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeHTTPRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
