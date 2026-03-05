
// Package examples provides usage examples for the njchilds90/go-pydantic-port library.
package examples

import (
	"context"
	"fmt"

	"github.com/njchilds90/go-pydantic-port/pkg/model"
)

// ModelExample demonstrates how to create and use a model.
func ExampleModel() {
	// Create a new model
	m := model.NewModel()

	// Set the model's properties
	if err := m.SetProperties(context.Background(), model.Properties{
		Name:  "John",
		Age:   30,
		Email: "john@example.com",
	}); err != nil {
		fmt.Printf("failed to set properties: %w", err)
		return
	}

	// Get the model's properties
	props, err := m.GetProperties(context.Background())
	if err != nil {
		fmt.Printf("failed to get properties: %w", err)
		return
	}

	// Print the model's properties
	fmt.Printf("properties: %+v", props)
}
// Output: properties: {Name:John Age:30 Email:john@example.com}
