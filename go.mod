module go-pydantic-port

// Package go-pydantic-port is a port of the popular Python library Pydantic for Go.
// It provides a simple and intuitive way to define data models and validate user input.

// go-pydantic-port is designed to be used in a variety of applications, from simple command-line tools to complex web services.

go 1.20

require (
	github.com/go-playground/validator/v10 v10.11.0
)

replace github.com/go-playground/validator/v10 => github.com/go-playground/validator/v10 v10.11.0