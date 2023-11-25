package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/onpremless/cli/ops"
	"github.com/onpremless/cli/tui/endpoint"
	api "github.com/onpremless/go-client"
	"github.com/spf13/cobra"
)

var endpointCmd = &cobra.Command{
	Use:   "endpoint",
	Short: "Endpoint API methods",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var endpointName string
var endpointLambdaID string

type endpointOps struct {
	ctx context.Context
}

func (op *endpointOps) Create(name string, path string, lambda string) tea.Cmd {
	return func() tea.Msg {
		endpt, err := ops.CreateEndpoint(op.ctx, &api.CreateEndpoint{
			Name:   name,
			Path:   path,
			Lambda: lambda,
		})

		return endpoint.EndpointCreateResponseMsg{
			Resp: &endpoint.EndpointCreateResponse{
				Endpoint: endpt,
				Err:      err,
			},
		}
	}
}

func (op *endpointOps) List() tea.Cmd {
	return func() tea.Msg {
		endpts, err := ops.ListEndpoints(op.ctx)

		return endpoint.EndpointListResponseMsg{
			Resp: &endpoint.EndpointListResponse{
				Endpoints: endpts,
				Err:       err,
			},
		}
	}
}

var endpointCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create",
	Run: func(cmd *cobra.Command, args []string) {
		var epoint string
		if len(args) > 0 {
			epoint = args[0]
		}

		var lambda *api.Lambda
		if endpointLambdaID != "" {
			var err error
			lambda, err = ops.GetLambda(cmd.Context(), endpointLambdaID)
			if err != nil {
				fmt.Printf("Failed to get lambda: %s", err)
				os.Exit(1)
			}
		}
		m := &endpoint.EndpointCreateModel{
			Name:            endpointName,
			Endpoint:        epoint,
			Lambda:          lambda,
			EndpointCreator: &endpointOps{ctx: cmd.Context()},
			LambdaLister:    &lambdaOps{ctx: cmd.Context()},
		}
		p := tea.NewProgram(endpoint.InitEndpointCreateModel(m))

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %s", err)
			os.Exit(1)
		}
	},
}

var endpointListCmd = &cobra.Command{
	Use:   "list",
	Short: "List",
	Run: func(cmd *cobra.Command, args []string) {
		m := &endpoint.EndpointListModel{
			Lister: &endpointOps{
				ctx: cmd.Context(),
			},
		}
		p := tea.NewProgram(endpoint.InitEndpointListModel(m))

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %s", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(endpointCmd)
	endpointCmd.AddCommand(endpointCreateCmd)
	endpointCmd.AddCommand(endpointListCmd)

	endpointCreateCmd.Flags().StringVarP(&endpointName, "name", "n", "", "endpoint name")
	endpointCreateCmd.Flags().StringVarP(&endpointLambdaID, "lambda-id", "l", "", "lambda id")
}
