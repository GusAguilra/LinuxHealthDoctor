package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/linuxhealthdoctor/lhd/internal/core"
	"github.com/linuxhealthdoctor/lhd/internal/dashboard"
	"github.com/linuxhealthdoctor/lhd/internal/plugin"
	"github.com/spf13/cobra"
)

func NewDashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Interactive TUI dashboard",
		Long:  `Launch the interactive terminal dashboard for real-time system health monitoring.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			m := dashboard.New()

			p := tea.NewProgram(m, tea.WithAltScreen())
			go func() {
				execCtx, err := plugin.NewExecutionContext(context.Background(), core.AllComponents(),
					plugin.WithTimeout(20*time.Second),
				)
				if err != nil {
					return
				}
				result := execCtx.Run()
				p.Send(dashboard.ResultsMsg{Result: result})
			}()

			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "dashboard error: %v\n", err)
				return err
			}
			return nil
		},
	}

	return cmd
}
