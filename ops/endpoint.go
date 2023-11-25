package ops

import (
	"context"
	"fmt"

	api "github.com/onpremless/go-client"
)

func CreateEndpoint(ctx context.Context, req *api.CreateEndpoint) (*api.Endpoint, error) {
	createResp, _, err := client.EndpointAPI.
		CreateEndpoint(ctx).
		CreateEndpoint(*req).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("error when calling `EndpointApi.CreateEndpoint``: %v", err)
	}

	return createResp, nil
}

func ListEndpoints(ctx context.Context) ([]api.Endpoint, error) {
	listResp, _, err := client.EndpointAPI.
		ListEndpoints(ctx).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("error when calling `EndpointApi.ListEndpoints``: %v", err)
	}

	return listResp, nil
}
