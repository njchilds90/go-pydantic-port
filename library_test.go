package pydantic

import (
    "testing"
)

func TestValidate(t *testing.T) {
    type User struct {
        Name  string `validate:"required"`
        Email string `validate:"required,email"`
    }

    user := User{
        Name:  "John Doe",
        Email: "johndoe@example.com",
    }

    err := Validate(user)
    if err != nil {
        t.Errorf("validation failed: %v", err)
    }
}

func TestModel(t *testing.T) {
    fields := map[string]interface{}{
        "name": "John Doe",
        "email": "johndoe@example.com",
    }

    model := Model("User", fields)
    if model == nil {
        t.Errorf("model is nil")
    }
}
