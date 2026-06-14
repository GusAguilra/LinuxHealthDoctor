package distro

func NewArch(info osRelease) Distro {
	return &GenericDistro{
		id:   info.ID,
		name: info.Name,
		vers: info.Version,
		pm: PackageManager{
			Name:       "pacman",
			Binary:     "/usr/bin/pacman",
			CheckCmd:   "pacman -Qu 2>/dev/null",
			InstallCmd: "pacman -S --noconfirm",
			UpdateCmd:  "pacman -Sy",
			ListCmd:    "pacman -Q",
		},
		sm: detectServiceManager(),
		fp: FilePaths{
			FSTab:         "/etc/fstab",
			DefaultLogDir: "/var/log",
			ReposDir:      "/etc/pacman.d",
			SSHConfig:     "/etc/ssh/sshd_config",
		},
	}
}
