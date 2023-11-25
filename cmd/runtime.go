package cmd

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/onpremless/cli/ops"
	"github.com/onpremless/cli/tui/runtime"
	"github.com/spf13/cobra"
)

var runtimeCmd = &cobra.Command{
	Use:   "runtime",
	Short: "Runtime API methods",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var runtimeName string

type runtimeOps struct {
	ctx context.Context
}

func (op *runtimeOps) Create(name string, path string) tea.Cmd {
	return func() tea.Msg {
		rt, err := ops.CreateRuntime(op.ctx, name, path)

		return runtime.RuntimeCreateResponseMsg{
			Resp: &runtime.RuntimeCreateResponse{
				Runtime: rt,
				Err:     err,
			},
		}
	}
}

func (op *runtimeOps) List() tea.Cmd {
	return func() tea.Msg {
		rt, err := ops.ListRuntimes(op.ctx)

		return runtime.RuntimeListResponseMsg{
			Resp: &runtime.RuntimeListResponse{
				Runtimes: rt,
				Err:      err,
			},
		}
	}
}

var runtimeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		m := &runtime.RuntimeCreateModel{
			Name: runtimeName,
			Path: args[0],
			Creator: &runtimeOps{
				ctx: cmd.Context(),
			},
		}
		p := tea.NewProgram(runtime.InitRuntimeCreateModel(m))

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %s", err)
		}
	},
}

var runtimeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List",
	Run: func(cmd *cobra.Command, args []string) {
		m := &runtime.RuntimeListModel{
			Lister: &runtimeOps{
				ctx: cmd.Context(),
			},
		}
		p := tea.NewProgram(runtime.InitRuntimeListModel(m))

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %s", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(runtimeCmd)
	runtimeCmd.AddCommand(runtimeCreateCmd)
	runtimeCmd.AddCommand(runtimeListCmd)

	runtimeCreateCmd.Flags().StringVarP(&runtimeName, "name", "n", "", "name")
}
