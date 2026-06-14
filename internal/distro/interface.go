package distro

type Distro interface {
	ID() string
	Name() string
	Version() string
	PackageManager() PackageManager
	ServiceManager() ServiceManager
	FilePaths() FilePaths
	KernelInfo() KernelInfo
	SecurityInfo() SecurityInfo
	Compatible(id string) bool
}

type PackageManager struct {
	Name       string
	Binary     string
	CheckCmd   string
	InstallCmd string
	UpdateCmd  string
	ListCmd    string
}

type ServiceManager struct {
	Name   string
	Binary string
}

type FilePaths struct {
	FSTab          string
	DefaultLogDir  string
	ReposDir       string
	SSHConfig      string
}

type KernelInfo struct {
	Version   string
	Arch      string
	BootTime  string
}

type SecurityInfo struct {
	SELinux      bool
	AppArmor     bool
	Firewall     string
}
