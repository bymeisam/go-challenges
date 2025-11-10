package main

type Vehicle interface {
	Drive() string
}

type Car struct{}

func (c Car) Drive() string {
	return "driving a car"
}

type Bike struct{}

func (b Bike) Drive() string {
	return "riding a bike"
}

func VehicleFactory(vehicleType string) Vehicle {
	switch vehicleType {
	case "car":
		return Car{}
	case "bike":
		return Bike{}
	default:
		return nil
	}
}

func main() {}
