// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logger(inner http.Handler) http.Handler {
	mw := func(w http.ResponseWriter, r *http.Request) {
		timeStarted := time.Now()
		inner.ServeHTTP(w, r)

		timeTaken := time.Since(timeStarted)
		log.Printf("%s %s took %v", r.Method, r.RequestURI, timeTaken)
	}
	return http.HandlerFunc(mw)
}
