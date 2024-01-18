package jsonclient

import "context"

type ClientOption[T any] func(*JSONClient[T]) error

func WithHeaders[T any](h map[string]string) ClientOption[T] {
	return func(y *JSONClient[T]) error {
		y.headers = h
		return nil
	}
}

func WithClient[T any](requester Requester) ClientOption[T] {

	return func(y *JSONClient[T]) error {
		y.client = requester
		return nil
	}
}

func WithContext[T any](ctx context.Context) ClientOption[T] {
	return func(y *JSONClient[T]) error {
		y.ctx = ctx
		return nil
	}
}

type QueryOption func(*QueryConfig) error
