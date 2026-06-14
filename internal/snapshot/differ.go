package snapshot

import (
	"fmt"
	"time"
)

type DiffType string

const (
	DiffAdded    DiffType = "added"
	DiffRemoved  DiffType = "removed"
	DiffModified DiffType = "modified"
)

type DiffEntry struct {
	Type      DiffType
	Path      string
	OldSize   int64
	NewSize   int64
	OldMode   string
	NewMode   string
	OldHash   string
	NewHash   string
}

type FileDiff struct {
	Added    []DiffEntry
	Removed  []DiffEntry
	Modified []DiffEntry
}

type ServiceDiff struct {
	Added    []DiffEntry
	Removed  []DiffEntry
	Modified []DiffEntry
}

type PackageDiff struct {
	Added    []PackageEntry
	Removed  []PackageEntry
	Upgraded []PackageUpgrade
}

type PackageUpgrade struct {
	Name       string
	OldVersion string
	NewVersion string
}

type SnapshotDiff struct {
	Timestamp     time.Time
	OldID         string
	NewID         string
	Files         FileDiff
	Services      ServiceDiff
	Packages      PackageDiff
	TotalChanges  int
}

func DiffSnapshots(old, new *Snapshot) (*SnapshotDiff, error) {
	if old == nil || new == nil {
		return nil, fmt.Errorf("old and new snapshots must not be nil")
	}

	diff := &SnapshotDiff{
		Timestamp: time.Now(),
		OldID:     old.ID,
		NewID:     new.ID,
	}

	if old.Manifest != nil && new.Manifest != nil {
		diff.Files = diffFiles(old.Manifest.Files, new.Manifest.Files)
		diff.Services = diffServices(old.Manifest.Services, new.Manifest.Services)
		diff.Packages = diffPackages(old.Manifest.Packages, new.Manifest.Packages)
	}

	diff.TotalChanges = len(diff.Files.Added) + len(diff.Files.Removed) + len(diff.Files.Modified) +
		len(diff.Services.Added) + len(diff.Services.Removed) + len(diff.Services.Modified) +
		len(diff.Packages.Added) + len(diff.Packages.Removed) + len(diff.Packages.Upgraded)

	return diff, nil
}

func diffFiles(old, new []FileEntry) FileDiff {
	oldMap := make(map[string]FileEntry)
	for _, f := range old {
		oldMap[f.Path] = f
	}

	newMap := make(map[string]FileEntry)
	for _, f := range new {
		newMap[f.Path] = f
	}

	var fd FileDiff

	for path, oldFile := range oldMap {
		if newFile, exists := newMap[path]; exists {
			if oldFile.Size != newFile.Size || oldFile.Mode != newFile.Mode || oldFile.SHA256 != newFile.SHA256 {
				fd.Modified = append(fd.Modified, DiffEntry{
					Type:      DiffModified,
					Path:      path,
					OldSize:   oldFile.Size,
					NewSize:   newFile.Size,
					OldMode:   fmt.Sprintf("%o", oldFile.Mode),
					NewMode:   fmt.Sprintf("%o", newFile.Mode),
					OldHash:   oldFile.SHA256,
					NewHash:   newFile.SHA256,
				})
			}
		} else {
			fd.Removed = append(fd.Removed, DiffEntry{
				Type:    DiffRemoved,
				Path:    path,
				OldSize: oldFile.Size,
				OldMode: fmt.Sprintf("%o", oldFile.Mode),
				OldHash: oldFile.SHA256,
			})
		}
	}

	for path, newFile := range newMap {
		if _, exists := oldMap[path]; !exists {
			fd.Added = append(fd.Added, DiffEntry{
				Type:    DiffAdded,
				Path:    path,
				NewSize: newFile.Size,
				NewMode: fmt.Sprintf("%o", newFile.Mode),
				NewHash: newFile.SHA256,
			})
		}
	}

	return fd
}

func diffServices(old, new []ServiceEntry) ServiceDiff {
	oldMap := make(map[string]ServiceEntry)
	for _, s := range old {
		oldMap[s.Name] = s
	}

	newMap := make(map[string]ServiceEntry)
	for _, s := range new {
		newMap[s.Name] = s
	}

	var sd ServiceDiff

	for name, oldSvc := range oldMap {
		if newSvc, exists := newMap[name]; exists {
			if oldSvc.Status != newSvc.Status {
				sd.Modified = append(sd.Modified, DiffEntry{
					Type: DiffModified,
					Path: name,
					OldHash: oldSvc.Status,
					NewHash: newSvc.Status,
				})
			}
		} else {
			sd.Removed = append(sd.Removed, DiffEntry{
				Type: DiffRemoved,
				Path: name,
				OldHash: oldSvc.Status,
			})
		}
	}

	for name, newSvc := range newMap {
		if _, exists := oldMap[name]; !exists {
			sd.Added = append(sd.Added, DiffEntry{
				Type: DiffAdded,
				Path: name,
				NewHash: newSvc.Status,
			})
		}
	}

	return sd
}

func diffPackages(old, new []PackageEntry) PackageDiff {
	oldMap := make(map[string]PackageEntry)
	for _, p := range old {
		oldMap[p.Name] = p
	}

	newMap := make(map[string]PackageEntry)
	for _, p := range new {
		newMap[p.Name] = p
	}

	var pd PackageDiff

	for name, oldPkg := range oldMap {
		if newPkg, exists := newMap[name]; exists {
			if oldPkg.Version != newPkg.Version {
				pd.Upgraded = append(pd.Upgraded, PackageUpgrade{
					Name:       name,
					OldVersion: oldPkg.Version,
					NewVersion: newPkg.Version,
				})
			}
		} else {
			pd.Removed = append(pd.Removed, oldPkg)
		}
	}

	for name, newPkg := range newMap {
		if _, exists := oldMap[name]; !exists {
			pd.Added = append(pd.Added, newPkg)
		}
	}

	return pd
}
