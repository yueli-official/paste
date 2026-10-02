// Command developer-api exports Paste's runtime PAT declaration.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/yueli-official/paste/internal/httpapi"
)

func main() {
	output := flag.String("output", "", "Output JSON file (required)")
	flag.Parse()
	if *output == "" {
		log.Fatal("--output is required")
	}
	data, err := json.MarshalIndent(httpapi.PersonalAPIDeclaration(), "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*output, append(data, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
}
