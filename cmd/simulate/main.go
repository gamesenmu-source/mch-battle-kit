// simulate runs locally, without a server, wallet or database.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	b "mch-light/engine/battle"
	"mch-light/engine/runner"
	"os"
)

func run() error {
	input := flag.String("input", "examples/battle.json", "battle input JSON")
	skillsPath := flag.String("skills", "data/skills.ndjson", "mechanics NDJSON")
	flag.Parse()
	file, err := os.Open(*skillsPath)
	if err != nil {
		return err
	}
	defer file.Close()
	skills, err := runner.LoadSkills(file)
	if err != nil {
		return err
	}
	requiredRaw, err := os.ReadFile("data/required-skills.json")
	if err != nil {
		return err
	}
	var required []uint32
	if err = json.Unmarshal(requiredRaw, &required); err != nil {
		return err
	}
	if err = runner.ValidateRequired(skills, required); err != nil {
		return err
	}
	b.RegisterRepository(&b.CatalogRepository{Skills: skills})
	raw, err := os.ReadFile(*input)
	if err != nil {
		return err
	}
	var in runner.Input
	if err = json.Unmarshal(raw, &in); err != nil {
		return err
	}
	result, err := runner.Run(in)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
