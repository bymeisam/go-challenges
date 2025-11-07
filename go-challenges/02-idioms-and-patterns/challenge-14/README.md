# Challenge 14: Functional Options Pattern

**Difficulty:** ⭐⭐⭐ Hard | **Time:** 25 min

Learn the functional options pattern - a Go idiom for flexible constructors.

## Tasks
Create a `Server` struct with optional configuration using functional options:
1. Define `Server` with Host, Port, Timeout fields
2. `NewServer(opts ...Option) *Server` - constructor with options
3. `WithHost(host string) Option` - option function
4. `WithPort(port int) Option` - option function

```bash
go test -v
```
