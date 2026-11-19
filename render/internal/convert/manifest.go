package convert

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Author is the author block of a plugin manifest.
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

type claudePlugin struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	License     string `json:"license,omitempty"`
	Author      Author `json:"author"`
}

type codexPlugin struct {
	claudePlugin
	Skills string    `json:"skills"`
	Hooks  *struct{} `json:"hooks,omitempty"`
}

type cursorPlugin struct {
	claudePlugin
	Skills   string    `json:"skills"`
	Agents   string    `json:"agents,omitempty"`
	Hooks    *struct{} `json:"hooks,omitempty"`
	Commands *[]string `json:"commands,omitempty"`
}

type codexEntry struct {
	Name   string `json:"name"`
	Source struct {
		Source string `json:"source"`
		Path   string `json:"path"`
	} `json:"source"`
	Policy struct {
		Installation   string `json:"installation"`
		Authentication string `json:"authentication"`
	} `json:"policy"`
	Category string `json:"category"`
}

type cursorEntry struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

// Manifests renders the Codex and Cursor manifests for every local plugin with skills in the Claude catalog under root, keyed by slash path relative to root.
func Manifests(root string) (map[string][]byte, error) {
	data, err := os.ReadFile(filepath.Join(root, ".claude-plugin", "marketplace.json"))
	if err != nil {
		return nil, err
	}
	var catalog struct {
		Name    string `json:"name"`
		Owner   Author `json:"owner"`
		Plugins []struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			Category    string          `json:"category"`
			Source      json.RawMessage `json:"source"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("parsing catalog: %w", err)
	}
	out := map[string][]byte{}
	codexMarket := struct {
		Name      string `json:"name"`
		Interface struct {
			DisplayName string `json:"displayName"`
		} `json:"interface"`
		Plugins []codexEntry `json:"plugins"`
	}{Name: catalog.Name, Plugins: []codexEntry{}}
	codexMarket.Interface.DisplayName = catalog.Name
	cursorMarket := struct {
		Name    string        `json:"name"`
		Owner   Author        `json:"owner"`
		Plugins []cursorEntry `json:"plugins"`
	}{Name: catalog.Name, Owner: catalog.Owner, Plugins: []cursorEntry{}}

	for _, p := range catalog.Plugins {
		var src string
		if json.Unmarshal(p.Source, &src) != nil || !strings.HasPrefix(src, "./") {
			continue
		}
		dir := filepath.Join(root, filepath.FromSlash(src))
		if !exists(filepath.Join(dir, "skills")) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "plugin.json"))
		if err != nil {
			return nil, err
		}
		var meta claudePlugin
		if err := json.Unmarshal(raw, &meta); err != nil {
			return nil, fmt.Errorf("parsing %s plugin.json: %w", p.Name, err)
		}
		hasHooks := exists(filepath.Join(dir, "hooks", "hooks.json"))

		codex := codexPlugin{claudePlugin: meta, Skills: "./skills/"}
		cursor := cursorPlugin{claudePlugin: meta, Skills: "./skills/"}
		if hasHooks {
			codex.Hooks, cursor.Hooks = &struct{}{}, &struct{}{}
		}
		if exists(filepath.Join(dir, "agents")) {
			cursor.Agents = "./agents/"
		}
		if exists(filepath.Join(dir, "commands")) {
			cursor.Commands = &[]string{}
		}
		if err := put(out, src+"/.codex-plugin/plugin.json", codex); err != nil {
			return nil, err
		}
		if err := put(out, src+"/.cursor-plugin/plugin.json", cursor); err != nil {
			return nil, err
		}

		var e codexEntry
		e.Name, e.Category = p.Name, p.Category
		e.Source.Source, e.Source.Path = "local", src
		e.Policy.Installation, e.Policy.Authentication = "AVAILABLE", "ON_INSTALL"
		codexMarket.Plugins = append(codexMarket.Plugins, e)
		cursorMarket.Plugins = append(cursorMarket.Plugins, cursorEntry{Name: p.Name, Source: src, Description: p.Description, Category: p.Category})
	}
	if len(codexMarket.Plugins) == 0 {
		return nil, errors.New("no local plugins with skills in the catalog")
	}
	if err := put(out, ".agents/plugins/marketplace.json", codexMarket); err != nil {
		return nil, err
	}
	if err := put(out, ".cursor-plugin/marketplace.json", cursorMarket); err != nil {
		return nil, err
	}
	return out, nil
}

func put(out map[string][]byte, path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	out[strings.TrimPrefix(path, "./")] = append(b, '\n')
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)
}
