package jsonclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type JSONClient[T any] struct {
	client       HTTPDoer
	ctx          context.Context
	baseUrl      *url.URL
	headers      map[string]string
	responseType T
}

type QueryConfig struct {
	params url.Values
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewClient[T any](baseUrlString string, options ...ClientOption[T]) (*JSONClient[T], error) {
	baseUrl, err := url.Parse(baseUrlString)
	if err != nil {
		return nil, fmt.Errorf("parsing url err : %w", err)
	}
	yc := &JSONClient[T]{baseUrl: baseUrl}
	for _, option := range options {
		option(yc)
	}
	if yc.client == nil {
		yc.client = http.DefaultClient
	}

	if yc.ctx == nil {
		yc.ctx = context.Background()
	}

	return yc, nil
}

// GET perform a get request that parses the incoming json to the client type
func (y *JSONClient[T]) GET(path string, options ...QueryOption) (T, error) {

	qc := &QueryConfig{}

	return y.makeHTTPRequest(path, "GET", qc.params, nil)
}

func (y *JSONClient[T]) POST(path string, obj T, options ...QueryOption) error {
	qc := &QueryConfig{}
	b, err := json.Marshal(obj)
	if err == nil {
		return err
	}
	_, err = y.makeHTTPRequest(path, "POST", qc.params, io.NopCloser(bytes.NewBuffer(b)))
	return err
}

func (y *JSONClient[T]) DELETE(path string, params url.Values, r io.Reader) (T, error) {

	return y.makeHTTPRequest(path, "DELETE", params, r)
}

func (y *JSONClient[T]) makeHTTPRequest(

	path string,
	httpMethod string,
	queryParameters url.Values,
	body io.Reader) (T, error) {
	u := y.baseUrl.JoinPath(path)

	if httpMethod == "GET" {
		q := u.Query()

		for k, v := range queryParameters {

			q.Set(k, strings.Join(v, ","))
		}

		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(y.ctx, httpMethod, u.String(), body)
	if err != nil {
		return y.responseType, fmt.Errorf("error: creating new request: %w", err)
	}

	for k, v := range y.headers {
		req.Header.Set(k, v)
	}

	log.Printf("%s %s\n", httpMethod, req.URL.String())

	res, err := y.client.Do(req)
	if err != nil {
		return y.responseType, err
	}

	if res == nil {
		return y.responseType, fmt.Errorf("error: calling %s returned empty response", u.String())
	}

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		return y.responseType, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return y.responseType, fmt.Errorf("error calling %s:\nstatus: %s\nresponseData: %s", u.String(), res.Status, responseData)
	}

	var responseObject T
	err = json.Unmarshal(responseData, &responseObject)

	if err != nil {

		return y.responseType, fmt.Errorf("error unmarshaling response: %+w", err)
	}

	return responseObject, nil
}
