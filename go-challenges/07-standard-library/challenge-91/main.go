package main

import "flag"

type Config struct {
	Host string
	Port int
	Debug bool
}

func ParseFlags() Config {
	host := flag.String("host", "localhost", "Server host")
	port := flag.Int("port", 8080, "Server port")
	debug := flag.Bool("debug", false, "Enable debug mode")
	
	flag.Parse()
	
	return Config{
		Host: *host,
		Port: *port,
		Debug: *debug,
	}
}

func main() {}
