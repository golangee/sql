// Copyright 2020 Torben Schinke
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sql

import (
	"context"
	"fmt"
)

type sqlCtxKey string

const (
	dbCtx sqlCtxKey = "db"
)

// WithContext creates a new context containing the given DBTX. Use this to implement orthogonal requirements
// like a scoped transaction. To learn more about when and how to use context, take a look at
// https://tip.golang.org/pkg/context/. It is a performance decision to reuse a repository, instead of creating
// the entire dependency chain for each request.
func WithContext(ctx context.Context, db DBTX) context.Context {
	return context.WithValue(ctx, dbCtx, db)
}

// FromContext is the counterpart of WithContext and returns a DBTX instance from the request-scoped context.
func FromContext(ctx context.Context) (DBTX, error) {
	if v := ctx.Value(dbCtx); v != nil {
		return v.(DBTX), nil
	}
	return nil, fmt.Errorf("context without DBTX, prepare with sql.WithContext")
}


