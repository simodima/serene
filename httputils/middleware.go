package httputils

import "net/http"

// Middleware represents a function that takes an http.Handler and returns
// another http.Handler. It is commonly used to implement behaviors such
// as logging, authentication, or request modification in a reusable way.
type Middleware func(http.Handler) http.Handler

// ApplyMiddlewares applies a chain of Middleware functions to an http.Handler.
// Starting with the provided base handler, each Middleware is applied in the
// order it appears in the middlewares slice. The final http.Handler returned
// will execute all the Middleware logic in sequence.
//
// Parameters:
//   - base: The initial http.Handler to which the Middleware functions will be applied.
//   - middlewares: A variadic list of Middleware functions to apply to the base handler.
//
// Returns:
//
//	A new http.Handler with all the Middleware functions applied in sequence.
//
// Example:
//
//	logger := func(next http.Handler) http.Handler {
//	    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        log.Println("Request received:", r.URL.Path)
//	        next.ServeHTTP(w, r)
//	    })
//	}
//
//	auth := func(next http.Handler) http.Handler {
//	    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        if r.Header.Get("Authorization") == "" {
//	            http.Error(w, "Forbidden", http.StatusForbidden)
//	            return
//	        }
//	        next.ServeHTTP(w, r)
//	    })
//	}
//
//	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    fmt.Fprintln(w, "Hello, world!")
//	})
//
//	wrapped := ApplyMiddlewares(base, logger, auth)
//	http.ListenAndServe(":8080", wrapped)
func ApplyMiddlewares(base http.Handler, middlewares ...Middleware) http.Handler {
	for _, h := range middlewares {
		base = h(base)
	}

	return base
}
