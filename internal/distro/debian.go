package distro

func NewDebian(info osRelease) Distro {
	sm := detectServiceManager()
	return &GenericDistro{
		id:   info.ID,
		name: info.Name,
		vers: info.Version,
		pm: PackageManager{
			Name:       "apt",
			Binary:     "/usr/bin/apt",
			CheckCmd:   "apt list --upgradable 2>/dev/null",
			InstallCmd: "apt install -y",
			UpdateCmd:  "apt update",
			ListCmd:    "dpkg -l",
		},
		sm: sm,
		fp: FilePaths{
			FSTab:         "/etc/fstab",
			DefaultLogDir: "/var/log",
			ReposDir:      "/etc/apt",
			SSHConfig:     "/etc/ssh/sshd_config",
		},
		si: SecurityInfo{
			AppArmor: true,
		},
	}
}
