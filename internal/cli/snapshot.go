package cli

import (
	"fmt"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/plugin"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/snapshot"
	"github.com/spf13/cobra"
)

func NewSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage system snapshots",
		Long:  `Create, list, diff, and manage point-in-time system snapshots.`,
	}

	cmd.AddCommand(NewSnapshotCreateCmd())
	cmd.AddCommand(NewSnapshotListCmd())
	cmd.AddCommand(NewSnapshotShowCmd())
	cmd.AddCommand(NewSnapshotDiffCmd())
	cmd.AddCommand(NewSnapshotCompareCmd())
	cmd.AddCommand(NewSnapshotDeleteCmd())

	return cmd
}

func NewSnapshotCreateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a system snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			execCtx, err := plugin.NewExecutionContext(cmd.Context(), core.AllComponents())
			if err != nil {
				return fmt.Errorf("execution: %w", err)
			}
			result := execCtx.Run()

			engine := snapshot.NewEngine(nil)
			snap, err := engine.Create(cmd.Context(), name, result)
			if err != nil {
				return fmt.Errorf("snapshot failed: %w", err)
			}

			fmt.Printf("Snapshot created: %s\n", snap.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "snapshot name")
	return cmd
}

func NewSnapshotListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Available snapshots:")
			fmt.Println("  (snapshot storage not yet connected)")
			return nil
		},
	}
}

func NewSnapshotShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [snapshot-id]",
		Short: "Show snapshot details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Snapshot: %s\n", args[0])
			return nil
		},
	}
}

func NewSnapshotDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff [snapshot1] [snapshot2]",
		Short: "Show differences between two snapshots",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Diff between %s and %s\n", args[0], args[1])
			return nil
		},
	}
}

func NewSnapshotCompareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "compare [snapshot-id]",
		Short: "Compare current state against a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Comparing current state against %s (not yet implemented)\n", args[0])
			return nil
		},
	}
}

func NewSnapshotDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [snapshot-id]",
		Short: "Delete a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Deleting snapshot: %s\n", args[0])
			return nil
		},
	}
}
