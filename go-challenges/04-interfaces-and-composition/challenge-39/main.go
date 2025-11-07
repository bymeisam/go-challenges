package main

type Base struct {
	ID   int
	Name string
}

func (b Base) GetID() int {
	return b.ID
}

type Extended struct {
	Base
	Extra string
}

func NewExtended(id int, name, extra string) Extended {
	return Extended{
		Base:  Base{ID: id, Name: name},
		Extra: extra,
	}
}

func main() {}
