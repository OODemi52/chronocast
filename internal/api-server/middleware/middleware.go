package middleware

import (
	"net/http"
)

func ChainMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {

	for _, mw := range middlewares {

		handler = mw(handler)

	}

	return handler

}
