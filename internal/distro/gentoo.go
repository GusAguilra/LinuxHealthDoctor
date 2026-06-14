package distro

func NewGentoo(info osRelease) Distro {
	return &GenericDistro{
		id:   info.ID,
		name: info.Name,
		vers: info.Version,
		pm: PackageManager{
			Name:       "portage",
			Binary:     "/usr/bin/emerge",
			CheckCmd:   "emerge --pretend --update @world 2>/dev/null",
			InstallCmd: "emerge",
			UpdateCmd:  "emerge --sync",
			ListCmd:    "equery list",
		},
		sm: detectServiceManager(),
		fp: FilePaths{
			FSTab:         "/etc/fstab",
			DefaultLogDir: "/var/log",
			ReposDir:      "/etc/portage",
			SSHConfig:     "/etc/ssh/sshd_config",
		},
	}
}
