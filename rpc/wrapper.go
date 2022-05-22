//Package rpc provides abstract rpc server
//
//Copyright (C) 2022 Alexander Kiryukhin <i@neonxp.dev>
//
//This file is part of go.neonxp.dev/jsonrpc2 project.
//
//This program is free software: you can redistribute it and/or modify
//it under the terms of the GNU General Public License as published by
//the Free Software Foundation, either version 3 of the License, or
//(at your option) any later version.
//
//This program is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//GNU General Public License for more details.
//
//You should have received a copy of the GNU General Public License
//along with this program.  If not, see <https://www.gnu.org/licenses/>.

package rpc

import (
	"context"
	"encoding/json"
)

func H[RQ any, RS any](handler func(context.Context, *RQ) (RS, error)) Handler {
	return func(ctx context.Context, in json.RawMessage) (json.RawMessage, error) {
		req := new(RQ)
		if err := json.Unmarshal(in, req); err != nil {
			return nil, ErrorFromCode(ErrCodeParseError)
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
