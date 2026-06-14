package distro

func NewAlpine(info osRelease) Distro {
	return &GenericDistro{
		id:   info.ID,
		name: info.Name,
		vers: info.Version,
		pm: PackageManager{
			Name:       "apk",
			Binary:     "/sbin/apk",
			CheckCmd:   "apk list --upgradable 2>/dev/null",
			InstallCmd: "apk add",
			UpdateCmd:  "apk update",
			ListCmd:    "apk list --installed",
		},
		sm: detectServiceManager(),
		fp: FilePaths{
			FSTab:         "/etc/fstab",
			DefaultLogDir: "/var/log",
			ReposDir:      "/etc/apk",
			SSHConfig:     "/etc/ssh/sshd_config",
		},
	}
}
