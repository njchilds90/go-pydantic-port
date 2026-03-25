package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunValidateAndSchema(t *testing.T) {
	d := t.TempDir()
	model := `{"name":"M","fields":[{"name":"id","type":"integer","required":true}]}`
	input := `{"id":1}`
	mf := filepath.Join(d, "model.json")
	inf := filepath.Join(d, "input.json")
	if err := os.WriteFile(mf, []byte(model), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inf, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runValidate([]string{"--model", mf, "--input", inf}); err != nil {
		t.Fatalf("runValidate err=%v", err)
	}
	if err := runSchema([]string{"--model", mf}); err != nil {
		t.Fatalf("runSchema err=%v", err)
	}
}
