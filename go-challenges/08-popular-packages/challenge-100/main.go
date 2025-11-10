package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type User struct {
	Email    string `validate:"required,email"`
	Age      int    `validate:"required,gte=18,lte=100"`
	Password string `validate:"required,min=8"`
	Website  string `validate:"omitempty,url"`
}

type Address struct {
	Street  string `validate:"required"`
	City    string `validate:"required"`
	Country string `validate:"required,iso3166_1_alpha2"`
	ZipCode string `validate:"required,numeric,len=5"`
}

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateUser(user *User) error {
	return validate.Struct(user)
}

func ValidateAddress(address *Address) error {
	return validate.Struct(address)
}

func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	
	if err == nil {
		return errors
	}
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			field := fieldError.Field()
			tag := fieldError.Tag()
			errors[field] = fmt.Sprintf("validation failed on '%s'", tag)
		}
	}
	
	return errors
}

func main() {}
