package jsonrpc2

import (
	"context"
	"encoding/json"
)

func Wrap[RQ any, RS any](handler func(context.Context, *RQ) (RS, error)) Handler {
	return func(ctx context.Context, in json.RawMessage) (json.RawMessage, error) {
		req := new(RQ)
		if err := json.Unmarshal(in, req); err != nil {
			return nil, NewError(ErrCodeParseError)
		}
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, Error{
				Code:    ErrUser,
				Message: err.Error(),
			}
		}
		return json.Marshal(resp)
	}
}

type Handler func(context.Context, json.RawMessage) (json.RawMessage, error)
