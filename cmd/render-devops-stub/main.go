// Command render-devops-stub builds the Lacrosse devops stub dashboard document
// using the same shipped model.BuildDevOpsStubDashboard + ToPostable path that
// terraform signoz_plugin_dashboard uses. Reads stub JSON files and writes the
// v2 dashboard document to stdout (or -o file).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SigNoz/terraform-provider-signoz/internal/plugindash"
)

func main() {
	stubDir := flag.String("stub-dir", "", "directory with fleet-nodes.json etc. (required)")
	out := flag.String("o", "", "output file (default stdout)")
	flag.Parse()
	if *stubDir == "" {
		fmt.Fprintln(os.Stderr, "usage: render-devops-stub -stub-dir path/to/stub-data [-o out.json]")
		os.Exit(2)
	}

	load := func(name string) map[string]interface{} {
		b, err := os.ReadFile(filepath.Join(*stubDir, name))
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", name, err)
			os.Exit(1)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			fmt.Fprintf(os.Stderr, "json %s: %v\n", name, err)
			os.Exit(1)
		}
		return m
	}

	pd := plugindash.BuildDevOpsStubDashboard(map[string]map[string]interface{}{
		"fleet":   load("fleet-nodes.json"),
		"status":  load("system-status.json"),
		"workers": load("worker-tally.json"),
		"otel":    load("otel-endpoints.json"),
		"pg":      load("pg-activity.json"),
	})
	doc, err := pd.DocumentJSON()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *out == "" {
		fmt.Println(doc)
		return
	}
	if err := os.WriteFile(*out, []byte(doc+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
