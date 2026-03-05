# Go Pydantic Port

A Go port of Python's Pydantic library, providing runtime data validation and parsing for Go developers.

## Overview

This library allows you to define and validate complex data structures with ease. It provides a simple and intuitive API for defining models, and supports a wide range of validation rules.

## Installation

To install this library, run the following command:

```go

go get github.com/go-pydantic-port
```

## Usage

Here is an example of how to define and validate a simple data model:

    package main

    import (
        "github.com/go-pydantic-port"
        "github.com/go-playground/validator/v10"
    )

    type User struct {
        Name  string `validate:"required"`
        Email string `validate:"required,email"`
    }

    func main() {
        user := User{
            Name:  "John Doe",
            Email: "johndoe@example.com",
        }

        err := pydantic.Validate(user)
        if err != nil {
            panic(err)
        }
    }

## API Reference

### pydantic.Validate

Validates a data model against a set of validation rules.

* `model`: The data model to validate
* `returns`: An error if the validation fails, or nil if the validation succeeds

### pydantic.Model

Defines a new data model.

* `name`: The name of the model
* `fields`: A map of field names to field definitions
* `returns`: A new data model