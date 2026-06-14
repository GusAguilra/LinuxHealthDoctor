package fleet

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Host struct {
	Hostname string            `yaml:"hostname"`
	Address  string            `yaml:"address"`
	Port     int               `yaml:"port"`
	User     string            `yaml:"user"`
	KeyFile  string            `yaml:"key_file"`
	Tags     []string          `yaml:"tags"`
	Labels   map[string]string `yaml:"labels"`
}

type Inventory struct {
	Hosts []Host `yaml:"hosts"`
}

func NewInventory() *Inventory {
	return &Inventory{Hosts: make([]Host, 0)}
}

func (inv *Inventory) AllHosts() []Host {
	return inv.Hosts
}

func LoadInventory(path string) (*Inventory, error) {
	if path == "" {
		return nil, fmt.Errorf("inventory path is empty")
	}

	expanded := path
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot resolve home dir: %w", err)
		}
		expanded = filepath.Join(home, path[1:])
	}

	data, err := os.ReadFile(expanded)
	if err != nil {
		return nil, fmt.Errorf("cannot read inventory file %s: %w", expanded, err)
	}

	var inv Inventory
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("cannot parse inventory file: %w", err)
	}

	for i := range inv.Hosts {
		if inv.Hosts[i].Address == "" {
			inv.Hosts[i].Address = inv.Hosts[i].Hostname
		}
	}

	return &inv, nil
}

func (inv *Inventory) Save(path string) error {
	expanded := path
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot resolve home dir: %w", err)
		}
		expanded = filepath.Join(home, path[1:])
	}

	data, err := yaml.Marshal(inv)
	if err != nil {
		return fmt.Errorf("cannot marshal inventory: %w", err)
	}

	dir := filepath.Dir(expanded)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create inventory dir %s: %w", dir, err)
	}

	if err := os.WriteFile(expanded, data, 0644); err != nil {
		return fmt.Errorf("cannot write inventory file %s: %w", expanded, err)
	}

	return nil
}

func (inv *Inventory) FilterByTags(tags []string) []Host {
	if len(tags) == 0 {
		return inv.Hosts
	}

	tagSet := make(map[string]bool)
	for _, t := range tags {
		tagSet[t] = true
	}

	var filtered []Host
	for _, host := range inv.Hosts {
		for _, tag := range host.Tags {
			if tagSet[tag] {
				filtered = append(filtered, host)
				break
			}
		}
	}

	return filtered
}

func (inv *Inventory) FilterByLabels(labels map[string]string) []Host {
	if len(labels) == 0 {
		return inv.Hosts
	}

	var filtered []Host
	for _, host := range inv.Hosts {
		match := true
		for k, v := range labels {
			if host.Labels[k] != v {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, host)
		}
	}

	return filtered
}

func (inv *Inventory) FilterByHostname(pattern string) []Host {
	var filtered []Host
	for _, host := range inv.Hosts {
		if matched, _ := filepath.Match(pattern, host.Hostname); matched {
			filtered = append(filtered, host)
		}
	}
	return filtered
}

func (inv *Inventory) AddHost(h Host) {
	for i, existing := range inv.Hosts {
		if existing.Hostname == h.Hostname {
			inv.Hosts[i] = h
			return
		}
	}
	inv.Hosts = append(inv.Hosts, h)
}

func (inv *Inventory) RemoveHost(hostname string) {
	for i, h := range inv.Hosts {
		if h.Hostname == hostname {
			inv.Hosts = append(inv.Hosts[:i], inv.Hosts[i+1:]...)
			return
		}
	}
}
