package distro

func NewNixOS(info osRelease) Distro {
	return &GenericDistro{
		id:   info.ID,
		name: info.Name,
		vers: info.Version,
		pm: PackageManager{
			Name:       "nix",
			Binary:     "/run/current-system/sw/bin/nix-env",
			CheckCmd:   "nix-env --upgrade --dry-run 2>/dev/null",
			InstallCmd: "nix-env -i",
			UpdateCmd:  "nix-channel --update",
			ListCmd:    "nix-env -q",
		},
		sm: detectServiceManager(),
		fp: FilePaths{
			FSTab:         "/etc/nixos/hardware-configuration.nix",
			DefaultLogDir: "/var/log",
			ReposDir:      "/etc/nixos",
			SSHConfig:     "/etc/ssh/sshd_config",
		},
	}
}
