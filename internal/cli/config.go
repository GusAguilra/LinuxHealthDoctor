package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `View, edit, and manage Linux Health Doctor configuration.`,
	}

	cmd.AddCommand(NewConfigInitCmd())
	cmd.AddCommand(NewConfigShowCmd())
	cmd.AddCommand(NewConfigSetCmd())
	cmd.AddCommand(NewConfigGetCmd())
	cmd.AddCommand(NewConfigValidateCmd())
	cmd.AddCommand(NewConfigPathCmd())
	cmd.AddCommand(NewConfigResetCmd())

	return cmd
}

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/etc/lhd/lhd.yaml"
	}
	return filepath.Join(home, ".config", "lhd", "lhd.yaml")
}

func NewConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create default configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getConfigPath()
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("create config dir: %w", err)
			}
			data, err := yaml.Marshal(config.Default())
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}
			if err := os.WriteFile(path, data, 0644); err != nil {
				return fmt.Errorf("write config: %w", err)
			}
			fmt.Printf("Default configuration created: %s\n", path)
			return nil
		},
	}
}

func NewConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFromContext(cmd.Context())
			if cfg == nil {
				fmt.Println("No configuration loaded")
				return nil
			}
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}
			fmt.Println(string(data))
			return nil
		},
	}
}

func NewConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Set %s = %s\n", args[0], args[1])
			return nil
		},
	}
}

func NewConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s = (value not loaded)\n", args[0])
			return nil
		},
	}
}

func NewConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getConfigPath()
			cfg, err := config.Load(path)
			if err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}
			fmt.Printf("Configuration valid: %s\n", path)
			yaml.NewEncoder(os.Stdout).Encode(cfg)
			return nil
		},
	}
}

func NewConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(getConfigPath())
			return nil
		},
	}
}

func NewConfigResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getConfigPath()
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove config: %w", err)
			}
			fmt.Printf("Configuration reset. Run 'lhd config init' to create defaults.\n")
			return nil
		},
	}
}
