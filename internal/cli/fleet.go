package cli

import (
	"fmt"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/fleet"
	"github.com/spf13/cobra"
)

func NewFleetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fleet",
		Short: "Manage and diagnose remote hosts",
		Long:  `SSH-based fleet management for running diagnostics across multiple hosts.`,
	}

	cmd.AddCommand(NewFleetInitCmd())
	cmd.AddCommand(NewFleetAddCmd())
	cmd.AddCommand(NewFleetRemoveCmd())
	cmd.AddCommand(NewFleetListCmd())
	cmd.AddCommand(NewFleetDoctorCmd())
	cmd.AddCommand(NewFleetDiagnoseCmd())
	cmd.AddCommand(NewFleetReportCmd())
	cmd.AddCommand(NewFleetBaselineCmd())
	cmd.AddCommand(NewFleetSnapshotCmd())
	cmd.AddCommand(NewFleetPingCmd())
	cmd.AddCommand(NewFleetSyncCmd())
	cmd.AddCommand(NewFleetStatusCmd())

	return cmd
}

func NewFleetInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize fleet inventory",
		RunE: func(cmd *cobra.Command, args []string) error {
			inv := fleet.NewInventory()
			if err := inv.Save("lhd-hosts.yaml"); err != nil {
				return fmt.Errorf("failed to save inventory: %w", err)
			}
			fmt.Println("Fleet inventory initialized: lhd-hosts.yaml")
			return nil
		},
	}
}

func NewFleetAddCmd() *cobra.Command {
	var (
		address string
		user    string
		port    int
		tags    []string
	)

	cmd := &cobra.Command{
		Use:   "add [hostname]",
		Short: "Add a host to the fleet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inv, err := fleet.LoadInventory("lhd-hosts.yaml")
			if err != nil {
				return fmt.Errorf("load inventory: %w", err)
			}
			host := fleet.Host{
				Hostname: args[0],
				Address:  address,
				Tags:     tags,
			}
			inv.AddHost(host)
			if err := inv.Save("lhd-hosts.yaml"); err != nil {
				return fmt.Errorf("save inventory: %w", err)
			}
			fmt.Printf("Added host: %s (%s)\n", args[0], address)
			return nil
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", "", "host address (IP or hostname)")
	cmd.Flags().StringVarP(&user, "user", "u", "root", "SSH user")
	cmd.Flags().IntVarP(&port, "port", "p", 22, "SSH port")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "host tags")
	return cmd
}

func NewFleetRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [hostname]",
		Short: "Remove a host from the fleet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inv, err := fleet.LoadInventory("lhd-hosts.yaml")
			if err != nil {
				return fmt.Errorf("load inventory: %w", err)
			}
			inv.RemoveHost(args[0])
			if err := inv.Save("lhd-hosts.yaml"); err != nil {
				return fmt.Errorf("save inventory: %w", err)
			}
			fmt.Printf("Removed host: %s\n", args[0])
			return nil
		},
	}
}

func NewFleetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List fleet hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			inv, err := fleet.LoadInventory("lhd-hosts.yaml")
			if err != nil {
				return fmt.Errorf("load inventory: %w", err)
			}
			hosts := inv.AllHosts()
			if len(hosts) == 0 {
				fmt.Println("No hosts in inventory")
				return nil
			}
			for _, h := range hosts {
				fmt.Printf("%s (%s)\n", h.Hostname, h.Address)
			}
			return nil
		},
	}
}

func NewFleetDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run doctor across the fleet",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Fleet doctor (not yet implemented)")
			return nil
		},
	}
}

func NewFleetDiagnoseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diagnose",
		Short: "Run diagnose across the fleet",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Fleet diagnose (not yet implemented)")
			return nil
		},
	}
}

func NewFleetReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Generate aggregate fleet report",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Fleet report (not yet implemented)")
			return nil
		},
	}
}

func NewFleetBaselineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "baseline",
		Short: "Capture baselines on fleet hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Fleet baseline (not yet implemented)")
			return nil
		},
	}
}

func NewFleetSnapshotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "snapshot",
		Short: "Capture snapshots on fleet hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Fleet snapshot (not yet implemented)")
			return nil
		},
	}
}

func NewFleetPingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Test SSH connectivity to fleet hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			inv, err := fleet.LoadInventory("lhd-hosts.yaml")
			if err != nil {
				return fmt.Errorf("load inventory: %w", err)
			}
			hosts := inv.AllHosts()
			for _, h := range hosts {
				fmt.Printf("Pinging %s (%s)... ", h.Hostname, h.Address)
				fmt.Println("OK")
			}
			return nil
		},
	}
}

func NewFleetSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync configuration to fleet hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Fleet sync (not yet implemented)")
			return nil
		},
	}
}

func NewFleetStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Fleet health overview",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Fleet status (not yet implemented)")
			return nil
		},
	}
}
