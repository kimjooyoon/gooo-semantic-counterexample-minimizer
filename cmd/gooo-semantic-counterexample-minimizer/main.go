package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-semantic-counterexample-minimizer/internal/minimizer"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "run" {
		fatal("usage: gooo-semantic-counterexample-minimizer run --meta PATH --contract PATH --source PATH --out DIR")
	}
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	meta := flags.String("meta", ".gooo/counterexample-minimizer.gooo", "authoritative semantic activity declaration")
	contract := flags.String("contract", "contracts/denominator-v1.json", "fixed denominator contract")
	source := flags.String("source", "", "semantic counterexample .gooo input")
	out := flags.String("out", "", "caller-owned empty output directory")
	if err := flags.Parse(os.Args[2:]); err != nil {
		os.Exit(2)
	}
	if *source == "" || *out == "" {
		fatal("--source and --out are required")
	}
	report, err := minimizer.RunWithMeta(*meta, *contract, *source, *out)
	if err != nil {
		fatal(err.Error())
	}
	data, err := json.Marshal(struct {
		Scenario string `json:"scenario"`
		State    string `json:"state"`
		Attempts int    `json:"attempts"`
	}{report.Scenario, report.State, report.Metrics.Attempts})
	if err != nil {
		fatal(err.Error())
	}
	fmt.Println(string(data))
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
