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
	"database/sql"
	"fmt"
	"net/http"
)

type KeyValues interface {
	ByName(name string) string
}

type Handler func(writer http.ResponseWriter, request *http.Request, params KeyValues) error

func WithTransaction(db *sql.DB) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request, params KeyValues) error {
			tx, err := db.BeginTx(r.Context(), nil)
			if err != nil {
				return err
			}
			txFinished := false
			defer func() {
				if !txFinished {
					// avoid memory leak
					_ = tx.Rollback()
				}
			}()
			newReq := r.WithContext(WithContext(r.Context(), tx))
			err = next(w, newReq, params)
			if err != nil {
				suppressedErr := tx.Rollback()
				if suppressedErr != nil {
					fmt.Println(suppressedErr)
				}
				txFinished = true
				return err
			}
			txFinished = true
			return tx.Commit()
		}
	}
}
