package distro

import (
	"bufio"
	"os"
	"strings"
)

type osRelease struct {
	ID      string
	IDLike  string
	Version string
	Name    string
}

func Detect() Distro {
	info := readOSRelease()

	if info.ID == "" {
		info = detectFallback()
	}

	switch info.ID {
	case "debian", "ubuntu", "linuxmint", "pop", "elementary":
		return NewDebian(info)
	case "rhel", "centos", "fedora", "rocky", "almalinux", "ol":
		return NewRHEL(info)
	case "arch", "manjaro", "endeavouros":
		return NewArch(info)
	case "opensuse", "opensuse-tumbleweed", "sles":
		return NewSUSE(info)
	case "gentoo":
		return NewGentoo(info)
	case "alpine":
		return NewAlpine(info)
	case "nixos":
		return NewNixOS(info)
	default:
		if strings.Contains(info.IDLike, "debian") {
			return NewDebian(info)
		}
		if strings.Contains(info.IDLike, "rhel") || strings.Contains(info.IDLike, "fedora") {
			return NewRHEL(info)
		}
		if strings.Contains(info.IDLike, "suse") {
			return NewSUSE(info)
		}
		return NewGeneric(info)
	}
}

func readOSRelease() osRelease {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return osRelease{}
	}
	defer f.Close()

	info := osRelease{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			info.ID = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		}
		if strings.HasPrefix(line, "ID_LIKE=") {
			info.IDLike = strings.Trim(strings.TrimPrefix(line, "ID_LIKE="), "\"")
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			info.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
		if strings.HasPrefix(line, "NAME=") {
			info.Name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
		}
	}
	return info
}

func detectFallback() osRelease {
	checks := []struct {
		path  string
		id    string
		name  string
	}{
		{"/etc/arch-release", "arch", "Arch Linux"},
		{"/etc/gentoo-release", "gentoo", "Gentoo"},
		{"/etc/alpine-release", "alpine", "Alpine Linux"},
		{"/etc/SuSE-release", "suse", "SUSE Linux"},
		{"/etc/fedora-release", "fedora", "Fedora"},
		{"/etc/centos-release", "centos", "CentOS"},
		{"/etc/redhat-release", "rhel", "Red Hat Enterprise Linux"},
		{"/etc/debian_version", "debian", "Debian GNU/Linux"},
	}

	for _, check := range checks {
		if _, err := os.Stat(check.path); err == nil {
			return osRelease{ID: check.id, Name: check.name}
		}
	}

	return osRelease{ID: "unknown", Name: "Unknown Linux"}
}
