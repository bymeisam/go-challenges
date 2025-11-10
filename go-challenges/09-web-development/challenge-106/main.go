package main

import (
	"net/http"
	"regexp"
	"strings"
)

type Route struct {
	pattern *regexp.Regexp
	handler http.HandlerFunc
	method  string
}

type Router struct {
	routes []Route
}

func NewRouter() *Router {
	return &Router{
		routes: make([]Route, 0),
	}
}

func (r *Router) Handle(method, pattern string, handler http.HandlerFunc) {
	// Convert pattern to regex
	// Example: /users/{id} -> /users/([^/]+)
	// Example: /posts/{postId}/comments/{commentId} -> /posts/([^/]+)/comments/([^/]+)
	regexPattern := pattern
	regexPattern = strings.ReplaceAll(regexPattern, "{", "(?P<")
	regexPattern = strings.ReplaceAll(regexPattern, "}", ">[^/]+)")
	regexPattern = "^" + regexPattern + "$"

	compiled := regexp.MustCompile(regexPattern)
	r.routes = append(r.routes, Route{
		pattern: compiled,
		handler: handler,
		method:  method,
	})
}

func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.Handle("GET", pattern, handler)
}

func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.Handle("POST", pattern, handler)
}

func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.Handle("PUT", pattern, handler)
}

func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.Handle("DELETE", pattern, handler)
}

func (r *Router) extractParams(pattern *regexp.Regexp, path string) map[string]string {
	matches := pattern.FindStringSubmatch(path)
	if matches == nil {
		return nil
	}

	params := make(map[string]string)
	names := pattern.SubexpNames()

	for i, name := range names {
		if i > 0 && i <= len(matches) {
			params[name] = matches[i]
		}
	}

	return params
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		if route.method != req.Method {
			continue
		}

		params := r.extractParams(route.pattern, req.URL.Path)
		if params != nil {
			// Add params to request context or query
			q := req.URL.Query()
			for k, v := range params {
				q.Set(k, v)
			}
			req.URL.RawQuery = q.Encode()

			route.handler(w, req)
			return
		}
	}

	http.NotFound(w, req)
}

func main() {}
