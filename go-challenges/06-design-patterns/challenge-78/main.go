package main

type HandlerFunc func(string) string

type Middleware func(HandlerFunc) HandlerFunc

func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(input string) string {
		result := next(input)
		return "Logged: " + result
	}
}

func AuthMiddleware(next HandlerFunc) HandlerFunc {
	return func(input string) string {
		if input == "unauthenticated" {
			return "Auth failed"
		}
		return next(input)
	}
}

func Chain(handler HandlerFunc, middlewares ...Middleware) HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func main() {}
