package examples

import (
	"context"
	"fmt"

	pydantic "github.com/njchilds90/go-pydantic-port"
)

func ExampleValidate() {
	type Output struct {
		Answer string `validate:"required,min=3"`
	}

	err := pydantic.Validate(context.Background(), Output{Answer: "yes"})
	fmt.Println(err == nil)
	// Output: true
}
