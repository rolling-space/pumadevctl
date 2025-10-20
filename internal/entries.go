package internal

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	Domain     string `json:"domain"`
	Mapping    string `json:"mapping"` // "" for symlink
	IsSymlink  bool   `json:"is_symlink"`
	LinkTarget string `json:"link_target,omitempty"`
}

func LoadEntries(dir string) ([]Entry, error) {
	items, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for _, de := range items {
		if de.IsDir() {
			// puma-dev uses hyphen to mimic nested hosts; actual subdirs are ignored
			continue
		}
		name := de.Name()
		full := filepath.Join(dir, name)
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			target, err := os.Readlink(full)
			if err != nil {
				target = fmt.Sprintf("error:%v", err)
			}
			entries = append(entries, Entry{
				Domain:     name,
				Mapping:    "",
				IsSymlink:  true,
				LinkTarget: target,
			})
			continue
		}
		data, err := os.ReadFile(full)
		mapping := ""
		if err != nil {
			mapping = fmt.Sprintf("error:%v", err)
		} else {
			mapping = strings.TrimSpace(string(data))
		}
		entries = append(entries, Entry{
			Domain:    name,
			Mapping:   mapping,
			IsSymlink: false,
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Domain < entries[j].Domain })
	return entries, nil
}

func ReadEntry(dir, domain string) (*Entry, error) {
	full := filepath.Join(dir, domain)
	info, err := os.Lstat(full)
	if err != nil {
		return nil, err
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		target, err := os.Readlink(full)
		if err != nil {
			return &Entry{Domain: domain, IsSymlink: true, LinkTarget: fmt.Sprintf("error:%v", err)}, nil
		}
		return &Entry{Domain: domain, IsSymlink: true, LinkTarget: target}, nil
	}
	b, err := os.ReadFile(full)
	if err != nil {
		return nil, err
	}
	return &Entry{Domain: domain, Mapping: strings.TrimSpace(string(b)), IsSymlink: false}, nil
}

func WriteEntry(dir, domain, mapping string, overwrite bool) error {
	if domain == "" {
		return errors.New("domain is required")
	}
	full := filepath.Join(dir, domain)
	if !overwrite {
		if _, err := os.Lstat(full); err == nil {
			return fmt.Errorf("entry %s already exists", domain)
		}
	}
	return os.WriteFile(full, []byte(mapping+""), 0644)
}

func CreateSymlink(dir, domain, target string, overwrite bool) error {
	if domain == "" {
		return errors.New("domain is required")
	}
	full := filepath.Join(dir, domain)
	if !overwrite {
		if _, err := os.Lstat(full); err == nil {
			return fmt.Errorf("entry %s already exists", domain)
		}
	} else {
		// remove any existing
		_ = os.Remove(full)
	}
	return os.Symlink(target, full)
}

func UpdateEntry(dir, domain, mapping string) error {
	full := filepath.Join(dir, domain)
	return os.WriteFile(full, []byte(mapping+""), 0644)
}

func UpdateSymlink(dir, domain, target string) error {
	full := filepath.Join(dir, domain)
	if err := os.Remove(full); err != nil {
		return err
	}
	return os.Symlink(target, full)
}

func DeleteEntry(dir, domain string) error {
	full := filepath.Join(dir, domain)
	return os.Remove(full)
}
