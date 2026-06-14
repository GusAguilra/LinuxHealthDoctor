package main

import (
	"os"

	_ "github.com/GusAguilra/LinuxHealthDoctor/internal/checks/cpu"
	_ "github.com/GusAguilra/LinuxHealthDoctor/internal/checks/disk"
	_ "github.com/GusAguilra/LinuxHealthDoctor/internal/checks/kernel"
	_ "github.com/GusAguilra/LinuxHealthDoctor/internal/checks/logs"
	_ "github.com/GusAguilra/LinuxHealthDoctor/internal/checks/memory"
	_ "github.com/GusAguilra/LinuxHealthDoctor/internal/checks/network"
	_ "github.com/GusAguilra/LinuxHealthDoctor/internal/checks/security"
	_ "github.com/GusAguilra/LinuxHealthDoctor/internal/checks/services"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/cli"
)

func main() {
	rootCmd := cli.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
