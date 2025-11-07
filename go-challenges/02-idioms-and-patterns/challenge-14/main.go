package main

type Server struct {
	Host    string
	Port    int
	Timeout int
}

type Option func(*Server)

func NewServer(opts ...Option) *Server {
	// TODO: Create server with defaults, apply options
	s := &Server{
		Host:    "localhost",
		Port:    8080,
		Timeout: 30,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithHost(host string) Option {
	// TODO: Return option function that sets host
	return func(s *Server) {
		s.Host = host
	}
}

func WithPort(port int) Option {
	// TODO: Return option function that sets port
	return func(s *Server) {
		s.Port = port
	}
}

func main() {}
