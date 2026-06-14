package distro

import "os/exec"

func detectServiceManager() ServiceManager {
	if _, err := exec.LookPath("systemctl"); err == nil {
		return ServiceManager{Name: "systemd", Binary: "systemctl"}
	}
	if _, err := exec.LookPath("rc-service"); err == nil {
		return ServiceManager{Name: "openrc", Binary: "rc-service"}
	}
	if _, err := exec.LookPath("runsvdir"); err == nil {
		return ServiceManager{Name: "runit", Binary: "runsvdir"}
	}
	if _, err := exec.LookPath("s6-svscan"); err == nil {
		return ServiceManager{Name: "s6", Binary: "s6-svscan"}
	}
	return ServiceManager{Name: "unknown", Binary: ""}
}

type GenericDistro struct {
	id    string
	name  string
	vers  string
	pm    PackageManager
	sm    ServiceManager
	fp    FilePaths
	ki    KernelInfo
	si    SecurityInfo
}

func NewGeneric(info osRelease) Distro {
	return &GenericDistro{
		id:   info.ID,
		name: info.Name,
		vers: info.Version,
		pm:   PackageManager{Name: "unknown", Binary: ""},
		sm:   detectServiceManager(),
		fp: FilePaths{
			FSTab:         "/etc/fstab",
			DefaultLogDir: "/var/log",
			SSHConfig:     "/etc/ssh/sshd_config",
		},
	}
}

func (d *GenericDistro) ID() string                       { return d.id }
func (d *GenericDistro) Name() string                     { return d.name }
func (d *GenericDistro) Version() string                  { return d.vers }
func (d *GenericDistro) PackageManager() PackageManager   { return d.pm }
func (d *GenericDistro) ServiceManager() ServiceManager   { return d.sm }
func (d *GenericDistro) FilePaths() FilePaths            { return d.fp }
func (d *GenericDistro) KernelInfo() KernelInfo          { return d.ki }
func (d *GenericDistro) SecurityInfo() SecurityInfo      { return d.si }
func (d *GenericDistro) Compatible(id string) bool        { return d.id == id }
