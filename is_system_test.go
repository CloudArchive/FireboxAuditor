package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestIsSystemJson(t *testing.T) {
	data, err := os.ReadFile("testdata/sample-config.xml")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseConfig(data)
	if err != nil {
		t.Fatal(err)
	}

	report := RunAudit(cfg)

	b, _ := json.MarshalIndent(report.Policies[0:2], "", "  ")
	fmt.Println(string(b))
}
