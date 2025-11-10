package main

type Handler interface {
	SetNext(handler Handler)
	Handle(request string) string
}

type BaseHandler struct {
	next Handler
}

func (h *BaseHandler) SetNext(handler Handler) {
	h.next = handler
}

type AuthHandler struct {
	BaseHandler
}

func (h *AuthHandler) Handle(request string) string {
	if request == "unauthenticated" {
		return "Auth failed"
	}
	if h.next != nil {
		return h.next.Handle(request)
	}
	return "Auth passed"
}

type ValidationHandler struct {
	BaseHandler
}

func (h *ValidationHandler) Handle(request string) string {
	if request == "invalid" {
		return "Validation failed"
	}
	if h.next != nil {
		return h.next.Handle(request)
	}
	return "Validation passed"
}

func main() {}
