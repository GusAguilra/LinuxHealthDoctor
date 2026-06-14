package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/monitor"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/storage"
	"github.com/spf13/cobra"
)

type monitorState struct {
	mu      sync.Mutex
	engine  *monitor.Engine
	store   *storage.SQLiteStore
}

var globalMonitor monitorState

func NewMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Continuous system monitoring",
		Long:  `Start, stop, and manage the continuous monitoring daemon.`,
	}

	cmd.AddCommand(NewMonitorStartCmd())
	cmd.AddCommand(NewMonitorStopCmd())
	cmd.AddCommand(NewMonitorStatusCmd())
	cmd.AddCommand(NewMonitorLogsCmd())
	cmd.AddCommand(NewMonitorAlertsCmd())
	cmd.AddCommand(NewMonitorThresholdsCmd())
	cmd.AddCommand(NewMonitorMetricsCmd())

	return cmd
}

func NewMonitorStartCmd() *cobra.Command {
	var interval time.Duration

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the monitoring daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalMonitor.mu.Lock()
			defer globalMonitor.mu.Unlock()
			if globalMonitor.engine != nil && globalMonitor.engine.IsRunning() {
				return fmt.Errorf("monitor is already running")
			}

			store, err := storage.NewSQLiteStore("/tmp/lhd-monitor.db")
			if err != nil {
				return fmt.Errorf("storage: %w", err)
			}
			eventBus := core.NewInMemoryEventBus()
			engine := monitor.NewEngine(store, eventBus)
			if interval == 0 {
				interval = 60 * time.Second
			}
			globalMonitor.engine = engine
			globalMonitor.store = store
			fmt.Printf("Monitoring daemon started (interval: %s)\n", interval)
			go func() {
				if err := engine.Start(context.Background(), interval); err != nil {
					fmt.Printf("Monitor error: %v\n", err)
				}
			}()
			return nil
		},
	}

	cmd.Flags().DurationVarP(&interval, "interval", "i", 60*time.Second, "collection interval")
	return cmd
}

func NewMonitorStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the monitoring daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalMonitor.mu.Lock()
			eng := globalMonitor.engine
			globalMonitor.mu.Unlock()
			if eng != nil {
				eng.Stop()
				fmt.Println("Monitoring daemon stopped")
			} else {
				fmt.Println("No monitoring daemon running")
			}
			return nil
		},
	}
}

func NewMonitorStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check monitoring daemon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalMonitor.mu.Lock()
			eng := globalMonitor.engine
			globalMonitor.mu.Unlock()
			if eng != nil && eng.IsRunning() {
				fmt.Println("Monitoring daemon: running")
			} else {
				fmt.Println("Monitoring daemon: stopped")
			}
			return nil
		},
	}
}

func NewMonitorLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "View monitoring logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Monitoring logs (not yet implemented)")
			return nil
		},
	}
}

func NewMonitorAlertsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "alerts",
		Short: "View and manage alerts",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalMonitor.mu.Lock()
			eng := globalMonitor.engine
			globalMonitor.mu.Unlock()
			if eng != nil {
				alerts := eng.Alerts()
				if len(alerts) == 0 {
					fmt.Println("No active alerts")
					return nil
				}
				for _, a := range alerts {
					fmt.Printf("[%s] %s\n", a.Severity, a.Message)
				}
			} else {
				fmt.Println("No active alerts")
			}
			return nil
		},
	}
}

func NewMonitorThresholdsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "thresholds",
		Short: "View and configure alert thresholds",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("CPU usage:    warning=80% critical=95%")
			fmt.Println("Memory usage: warning=80% critical=95%")
			fmt.Println("Disk usage:   warning=80% critical=92%")
			return nil
		},
	}
}

func NewMonitorMetricsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "metrics",
		Short: "View collected metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Collected metrics (not yet implemented)")
			return nil
		},
	}
}
