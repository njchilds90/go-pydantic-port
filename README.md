# Go Pydantic Port

A Go port of Python's Pydantic library, providing runtime data validation and parsing for Go developers.

## Overview

This library allows you to define and validate complex data structures with ease. It provides a simple and intuitive API for defining models, and supports a wide range of validation rules.

## Installation

To install this library, run the following command:

```bash
go get github.com/njchilds90/go-pydantic-port
```

## Usage

Here is an example of how to define and validate a simple data model:

```go
package main

import (
	"context"
	"fmt"

	"github.com/go-pydantic-port"
	"github.com/go-playground/validator/v10"
)

// User represents a simple data model
//godoc
func NewUser(name string, email string) (*User, error) {
	if name == "" || email == "" {
		return nil, fmt.Errorf("invalid input: name and email are required")
	}
	return &User{
		Name:  name,
		Email: email,
	}, nil
}

type User struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func main() {
	ctx := context.Background()
	user, err := NewUser("John Doe", "johndoe@example.com")
	if err != nil {
		panic(err)
	}
	
	// example of using godoc comments for exported functions
	// Validate takes a User and returns an error if validation fails
	// godoc
	func Validate(ctx context.Context, u *User) error {
		if u == nil {
			return fmt.Errorf("invalid input: user is nil")
		}
		return pydantic.Validate(ctx, u)
	}
	
	if err := Validate(ctx, user); err != nil {
		panic(err)
	}
}

## API Reference

### pydantic.Validate

Validates a data model against a set of validation rules.
* `ctx`: the context for the validation
* `model`: The data model to validate
* `returns`: An error if the validation fails, or nil if the validation succeeds

### pydantic.Model

Defines a new data model.
* `name`: The name of the model
* `fields`: A map of field names to field definitions
* `returns`: A new data model

## pkg.go.dev badge

[![PkgGoDev](https://pkg.go.dev/badge/github.com/njchilds90/go-pydantic-port)](https://pkg.go.dev/github.com/njchilds90/go-pydantic-port)
