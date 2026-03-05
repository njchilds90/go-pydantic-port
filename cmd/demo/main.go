package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/njchilds90/go-pydantic-port"
)

func main() {
	pyPortPtr := flag.String("py", "", "path to python port")
	flag.Parse()

	ctx := context.Background()
	if err := goPydanticPort.Run(ctx, *pyPortPtr); err != nil {
		log.Printf("error running demo: %v\n", err)
		os.Exit(1)
	}
}
