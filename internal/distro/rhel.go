package distro

import "os/exec"

func NewRHEL(info osRelease) Distro {
	sm := detectServiceManager()
	binary := "/usr/bin/dnf"
	if _, err := exec.LookPath("dnf"); err != nil {
		binary = "/usr/bin/yum"
	}
	return &GenericDistro{
		id:   info.ID,
		name: info.Name,
		vers: info.Version,
		pm: PackageManager{
			Name:       "dnf",
			Binary:     binary,
			CheckCmd:   binary + " check-update -q 2>/dev/null",
			InstallCmd: binary + " install -y",
			UpdateCmd:  binary + " makecache",
			ListCmd:    "rpm -qa",
		},
		sm: sm,
		fp: FilePaths{
			FSTab:         "/etc/fstab",
			DefaultLogDir: "/var/log",
			ReposDir:      "/etc/yum.repos.d",
			SSHConfig:     "/etc/ssh/sshd_config",
		},
		si: SecurityInfo{
			SELinux: true,
		},
	}
}
