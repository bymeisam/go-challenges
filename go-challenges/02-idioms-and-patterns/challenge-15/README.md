# Challenge 15: Constructor Pattern

**Difficulty:** ⭐⭐ Medium | **Time:** 15 min

Learn Go's constructor pattern (no `new` keyword like JS!).

## Tasks
1. Define `Database` struct with conn string
2. `NewDatabase(conn string) (*Database, error)` - constructor with validation
3. `MustNewDatabase(conn string) *Database` - panic on error (use sparingly!)

```bash
go test -v
```
