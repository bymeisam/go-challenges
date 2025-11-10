package main

type House struct {
	Windows int
	Doors   int
	Floors  int
	HasGarage bool
}

type HouseBuilder struct {
	house House
}

func NewHouseBuilder() *HouseBuilder {
	return &HouseBuilder{}
}

func (b *HouseBuilder) Windows(n int) *HouseBuilder {
	b.house.Windows = n
	return b
}

func (b *HouseBuilder) Doors(n int) *HouseBuilder {
	b.house.Doors = n
	return b
}

func (b *HouseBuilder) Floors(n int) *HouseBuilder {
	b.house.Floors = n
	return b
}

func (b *HouseBuilder) Garage(hasGarage bool) *HouseBuilder {
	b.house.HasGarage = hasGarage
	return b
}

func (b *HouseBuilder) Build() House {
	return b.house
}

func main() {}
