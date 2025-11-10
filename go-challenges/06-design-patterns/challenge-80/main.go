package main

import "sync"

type Connection struct {
	ID int
}

type Pool struct {
	mu    sync.Mutex
	items []*Connection
	size  int
}

func NewPool(size int) *Pool {
	pool := &Pool{
		items: make([]*Connection, 0, size),
		size:  size,
	}
	for i := 0; i < size; i++ {
		pool.items = append(pool.items, &Connection{ID: i})
	}
	return pool
}

func (p *Pool) Acquire() *Connection {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if len(p.items) == 0 {
		return nil
	}
	
	conn := p.items[0]
	p.items = p.items[1:]
	return conn
}

func (p *Pool) Release(conn *Connection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.items = append(p.items, conn)
}

func main() {}
