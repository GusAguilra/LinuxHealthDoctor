package snapshot

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

type ManifestGenerator struct {
	IncludePatterns []string
	ExcludePatterns []string
}

func NewManifestGenerator(includes, excludes []string) *ManifestGenerator {
	return &ManifestGenerator{
		IncludePatterns: includes,
		ExcludePatterns: excludes,
	}
}

func (mg *ManifestGenerator) Generate(ctx context.Context, root string) (*Manifest, error) {
	manifest := &Manifest{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if !mg.matchesInclude(rel) {
			return nil
		}
		if mg.matchesExclude(rel) {
			return nil
		}

		checksum, err := fileChecksum(path)
		if err != nil {
			return err
		}

		manifest.Files = append(manifest.Files, FileEntry{
			Path:   rel,
			Size:   info.Size(),
			Mode:   info.Mode(),
			SHA256: checksum,
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk error: %w", err)
	}

	return manifest, nil
}

func (mg *ManifestGenerator) matchesInclude(path string) bool {
	if len(mg.IncludePatterns) == 0 {
		return true
	}
	for _, pattern := range mg.IncludePatterns {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func (mg *ManifestGenerator) matchesExclude(path string) bool {
	for _, pattern := range mg.ExcludePatterns {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func fileChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum), nil
}

func (m *Manifest) AddPackage(name, version string) {
	m.Packages = append(m.Packages, PackageEntry{
		Name:    name,
		Version: version,
	})
}

func (m *Manifest) AddService(name, status string) {
	m.Services = append(m.Services, ServiceEntry{
		Name:   name,
		Status: status,
	})
}

func (m *Manifest) TotalFiles() int {
	return len(m.Files)
}

func (m *Manifest) TotalPackages() int {
	return len(m.Packages)
}

func (m *Manifest) TotalServices() int {
	return len(m.Services)
}

func (m *Manifest) FileSize() int64 {
	var total int64
	for _, f := range m.Files {
		total += f.Size
	}
	return total
}
