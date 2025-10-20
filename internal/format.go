package internal

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.AppendHeader(table.Row{"Mapping", "Domains", "Note"})
	groups := GroupByMapping(entries)
	for _, g := range groups {
		header := g.Mapping
		if header == "" {
			header = "(empty)"
		}
		domains := strings.Join(g.Domains, ", ")
		// Style mapping header in cyan; note in yellow (if present)
		styledHeader := text.FgCyan.Sprint(header)
		styledNote := g.Note
		if g.Note != "" {
			styledNote = text.FgYellow.Sprint(g.Note)
		}
		tw.AppendRow(table.Row{styledHeader, domains, styledNote})
	}
	tw.SetStyle(table.StyleRounded)
	tw.Style().Format.Header = text.FormatDefault
	tw.Render()
}

func PrintListJSON(entries []Entry) error {
	groups := GroupByMapping(entries)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(groups)
}
