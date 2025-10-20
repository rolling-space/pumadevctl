package internal

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
)

type ListGroup struct {
	Mapping string   `json:"mapping"` // "(symlink)" or concrete mapping
	Domains []string `json:"domains"`
	Note    string   `json:"note,omitempty"`
}

func GroupByMapping(entries []Entry) []ListGroup {
	buckets := map[string][]string{}
	symlinkKey := "(symlink)"
	for _, e := range entries {
		key := e.Mapping
		if e.IsSymlink {
			key = symlinkKey
		}
		buckets[key] = append(buckets[key], e.Domain)
	}
	groups := make([]ListGroup, 0, len(buckets))
	for k, v := range buckets {
		sort.Strings(v)
		note := ""
		if k != symlinkKey && len(v) > 1 {
			note = "duplicate mapping"
		}
		groups = append(groups, ListGroup{Mapping: k, Domains: v, Note: note})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Mapping < groups[j].Mapping })
	return groups
}

func PrintListFancy(entries []Entry) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s", bold("Puma-dev mappings"))
	fmt.Println(strings.Repeat("─", 60))
	groups := GroupByMapping(entries)
	for _, g := range groups {
		header := g.Mapping
		if header == "" {
			header = "(empty)"
		}
		if g.Note != "" {
			fmt.Printf("%s  %s", color.CyanString(header), color.YellowString("[%s]", g.Note))
		} else {
			fmt.Printf("%s", color.CyanString(header))
		}
		for _, d := range g.Domains {
			fmt.Printf("  • %s", d)
		}
		fmt.Println()
	}
}

func PrintListJSON(entries []Entry) error {
	groups := GroupByMapping(entries)
	enc := json.NewEncoder(color.Output)
	enc.SetIndent("", "  ")
	return enc.Encode(groups)
}
