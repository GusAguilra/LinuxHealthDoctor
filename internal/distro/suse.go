package distro

func NewSUSE(info osRelease) Distro {
	return &GenericDistro{
		id:   info.ID,
		name: info.Name,
		vers: info.Version,
		pm: PackageManager{
			Name:       "zypper",
			Binary:     "/usr/bin/zypper",
			CheckCmd:   "zypper list-updates 2>/dev/null",
			InstallCmd: "zypper install -y",
			UpdateCmd:  "zypper refresh",
			ListCmd:    "rpm -qa",
		},
		sm: detectServiceManager(),
		fp: FilePaths{
			FSTab:         "/etc/fstab",
			DefaultLogDir: "/var/log",
			ReposDir:      "/etc/zypp/repos.d",
			SSHConfig:     "/etc/ssh/sshd_config",
		},
	}
}
