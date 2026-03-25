package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	pydantic "github.com/njchilds90/go-pydantic-port"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "validate":
		err = runValidate(os.Args[2:])
	case "schema":
		err = runSchema(os.Args[2:])
	case "serve":
		err = runServe(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("usage: pydantic <validate|schema|serve> [flags]")
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	modelFile := fs.String("model", "model.json", "model definition file")
	inputFile := fs.String("input", "input.json", "input payload file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	m, err := loadModel(*modelFile)
	if err != nil {
		return err
	}
	in := map[string]any{}
	raw, err := os.ReadFile(*inputFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return err
	}
	if err := pydantic.ValidateMap(context.Background(), m, in); err != nil {
		return err
	}
	fmt.Println("valid")
	return nil
}

func runSchema(args []string) error {
	fs := flag.NewFlagSet("schema", flag.ContinueOnError)
	modelFile := fs.String("model", "model.json", "model definition file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	m, err := loadModel(*modelFile)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(m.Schema(), "", "  ")
	fmt.Println(string(out))
	return nil
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	modelFile := fs.String("model", "model.json", "model definition file")
	addr := fs.String("addr", ":8080", "listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	m, err := loadModel(*modelFile)
	if err != nil {
		return err
	}
	h := http.NewServeMux()
	h.HandleFunc("/schema", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(m.Schema())
	})
	h.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		in := map[string]any{}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := pydantic.ValidateMap(r.Context(), m, in); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := http.ListenAndServe(*addr, h); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func loadModel(path string) (*pydantic.Model, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m pydantic.Model
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
