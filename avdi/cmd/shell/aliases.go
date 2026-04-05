package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type aliasStore struct {
	path string
	m    map[string]string
}

func newAliasStore() *aliasStore {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	p := filepath.Join(dir, "avdi", "aliases.json")
	s := &aliasStore{path: p, m: make(map[string]string)}
	s.load()
	return s
}

func (s *aliasStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &s.m)
	if s.m == nil {
		s.m = make(map[string]string)
	}
}

func (s *aliasStore) save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *aliasStore) apply(line string) string {
	args := splitArgs(line)
	if len(args) == 0 {
		return line
	}
	if repl, ok := s.m[args[0]]; ok {
		if len(args) == 1 {
			return repl
		}
		return repl + " " + strings.Join(args[1:], " ")
	}
	return line
}

func cmdAlias(s *aliasStore, rest string) error {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		if len(s.m) == 0 {
			fmt.Println("No aliases. Use: alias name=command")
			return nil
		}
		var keys []string
		for k := range s.m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %s = %s\n", k, s.m[k])
		}
		return nil
	}
	if !strings.Contains(rest, "=") {
		if v, ok := s.m[rest]; ok {
			fmt.Printf("%s = %s\n", rest, v)
			return nil
		}
		return fmt.Errorf("unknown alias: %s", rest)
	}
	parts := strings.SplitN(rest, "=", 2)
	name := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	if name == "" {
		return fmt.Errorf("empty alias name")
	}
	s.m[name] = val
	if err := s.save(); err != nil {
		return fmt.Errorf("save aliases: %w", err)
	}
	printSuccess("alias set: " + name)
	return nil
}

func cmdUnalias(s *aliasStore, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("usage: unalias <name>")
	}
	if _, ok := s.m[name]; !ok {
		return fmt.Errorf("unknown alias: %s", name)
	}
	delete(s.m, name)
	if err := s.save(); err != nil {
		return fmt.Errorf("save aliases: %w", err)
	}
	printSuccess("removed alias: " + name)
	return nil
}
