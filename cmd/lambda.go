package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	api "github.com/onpremless/go-client"
	"github.com/onpremless/opcli/ops"
	"github.com/onpremless/opcli/tui/lambda"
	"github.com/spf13/cobra"
)

var lambdaCmd = &cobra.Command{
	Use:   "lambda",
	Short: "Lambda API methods",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var lambdaName string
var lambdaRuntime string
var lambdaType string

type lambdaOps struct {
	ctx context.Context
}

func (op *lambdaOps) Create(name string, runtime string, lambdaType string, path string) tea.Cmd {
	return func() tea.Msg {
		l, err := ops.CreateLambda(op.ctx, ops.CreateLambdaM{
			Name:       name,
			Runtime:    runtime,
			LambdaType: lambdaType,
		}, path)

		return lambda.LambdaCreateResponseMsg{
			Resp: &lambda.LambdaCreateResponse{
				Lambda: l,
				Err:    err,
			},
		}
	}
}

func (op *lambdaOps) List() tea.Cmd {
	return func() tea.Msg {
		lambdas, err := ops.ListLambdas(op.ctx)

		return lambda.LambdaListResponseMsg{
			Resp: &lambda.LambdaListResponse{
				Lambdas: lambdas,
				Err:     err,
			},
		}
	}
}

func (op *lambdaOps) Start(id string) tea.Cmd {
	return func() tea.Msg {
		l, err := ops.StartLambda(op.ctx, id)

		return lambda.LambdaStartResponseMsg{
			Resp: &lambda.LambdaStartResponse{
				Lambda: l,
				Err:    err,
			},
		}
	}
}

func (op *lambdaOps) Destroy(id string) tea.Cmd {
	return func() tea.Msg {
		err := ops.DestroyLambda(op.ctx, id)

		return lambda.LambdaDestroyResponseMsg{
			Resp: &lambda.LambdaDestroyResponse{
				Err: err,
			},
		}
	}
}

func lambdaCreateProgram(cmd *cobra.Command, args []string) *tea.Program {
	var runtime *api.Runtime
	if lambdaRuntime != "" {
		var err error
		runtime, err = ops.GetRuntime(cmd.Context(), lambdaRuntime)
		if err != nil {
			fmt.Printf("Failed to get runtime: %s", err)
			os.Exit(1)
		}
	}

	m := &lambda.LambdaCreateModel{
		Name:          lambdaName,
		Runtime:       runtime,
		LambdaType:    lambdaType,
		Path:          args[0],
		LambdaCreator: &lambdaOps{ctx: cmd.Context()},
		RuntimeLister: &runtimeOps{ctx: cmd.Context()},
	}

	return tea.NewProgram(lambda.InitLambdaCreateModel(m))
}

var lambdaCreateCmd = &cobra.Command{
	Use:   "create [path]",
	Short: "Create new lambda",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p := lambdaCreateProgram(cmd, args)

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %s", err)
			os.Exit(1)
		}
	},
}

var lambdaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List lambdas",
	Run: func(cmd *cobra.Command, args []string) {
		m := &lambda.LambdaListModel{
			Lister: &lambdaOps{
				ctx: cmd.Context(),
			},
		}
		p := tea.NewProgram(lambda.InitLambdaListModel(m))

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %s", err)
			os.Exit(1)
		}
	},
}

var lambdaStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start lambda",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		m := &lambda.LambdaStartModel{
			LambdaID: args[0],
			Starter:  &lambdaOps{ctx: cmd.Context()},
		}

		p := tea.NewProgram(lambda.InitLambdaStartModel(m))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %s", err)
			os.Exit(1)
		}
	},
}

var lambdaDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy lambda, aka create + start",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p := lambdaCreateProgram(cmd, args)
		m, err := p.Run()
		if err != nil {
			fmt.Printf("Error: %s", err)
			os.Exit(1)
		}

		cm := m.(*lambda.LambdaCreateModel)
		l := cm.GetLambda()
		if l == nil {
			// The error has already been printed as part of the previous program output
			os.Exit(1)
		}

		sm := &lambda.LambdaStartModel{
			LambdaID: cm.GetLambda().Id,
			Starter:  &lambdaOps{ctx: cmd.Context()},
		}

		p = tea.NewProgram(lambda.InitLambdaStartModel(sm))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %s", err)
			os.Exit(1)
		}
	},
}

var lambdaDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy lambda",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		m := &lambda.LambdaDestroyModel{
			LambdaID:  args[0],
			Destroyer: &lambdaOps{ctx: cmd.Context()},
		}

		p := tea.NewProgram(lambda.InitLambdaDestroyModel(m))
		if err := p.Start(); err != nil {
			fmt.Printf("Error: %s", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(lambdaCmd)
	lambdaCmd.AddCommand(lambdaDeployCmd)
	lambdaCmd.AddCommand(lambdaCreateCmd)
	lambdaCmd.AddCommand(lambdaListCmd)
	lambdaCmd.AddCommand(lambdaStartCmd)
	lambdaCmd.AddCommand(lambdaDestroyCmd)

	lambdaCreateCmd.Flags().StringVarP(&lambdaName, "name", "n", "", "name")
	lambdaCreateCmd.Flags().StringVarP(&lambdaRuntime, "runtime", "r", "", "runtime")
	lambdaCreateCmd.Flags().StringVarP(&lambdaType, "type", "t", "", "type of lambda (ENDPOINT | INTERNAL)")

	lambdaDeployCmd.Flags().StringVarP(&lambdaName, "name", "n", "", "name")
	lambdaDeployCmd.Flags().StringVarP(&lambdaRuntime, "runtime", "r", "", "runtime")
	lambdaDeployCmd.Flags().StringVarP(&lambdaType, "type", "e", "", "type of lambda (ENDPOINT | INTERNAL)")
}
