package ops

import (
	"context"
	"encoding/json"
	"fmt"

	api "github.com/onpremless/go-client"
)

type CreateLambdaM struct {
	Name       string
	Runtime    string
	LambdaType string
}

func CreateLambda(ctx context.Context, lambda CreateLambdaM, path string) (*api.Lambda, error) {
	uploadID, err := upload(ctx, path, true)
	if err != nil {
		return nil, err
	}

	createResp, r, err := client.LambdaAPI.
		CreateLambda(ctx).
		CreateLambda(api.CreateLambda{
			Name:       lambda.Name,
			Runtime:    lambda.Runtime,
			LambdaType: lambda.LambdaType,
			Archive:    uploadID,
		}).
		Execute()
	if err != nil {
		var details api.Error
		json.NewDecoder(r.Body).Decode(&details)
		return nil, fmt.Errorf("error when calling `LambdaApi.CreateLambda``: %v\n%v", err, details.GetError())
	}

	return createResp, nil
}

func GetLambda(ctx context.Context, id string) (*api.Lambda, error) {
	resp, _, err := client.LambdaAPI.
		GetLambda(ctx, id).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("error when calling `LambdaApi.GetLambda``: %v", err)
	}

	return resp, nil
}

func ListLambdas(ctx context.Context) ([]api.Lambda, error) {
	resp, _, err := client.LambdaAPI.
		ListLambdas(ctx).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("error when calling `LambdaApi.ListLambdas``: %v", err)
	}

	return resp, nil
}

func StartLambda(ctx context.Context, id string) (*api.Lambda, error) {
	taskResp, _, err := client.LambdaAPI.
		StartLambda(ctx, id).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("error when calling `LambdaApi.StartLambda``: %v", err)
	}

	res, err := pollTask(ctx, taskResp.GetTask())
	if err != nil {
		return nil, err
	}

	if res.GetStatus() == "FAILED" {
		return nil, fmt.Errorf("error when starting lambda: %v", res.GetDetails()["error"])
	}

	return GetLambda(ctx, id)
}

func DestroyLambda(ctx context.Context, id string) error {
	taskResp, _, err := client.LambdaAPI.
		DestroyLambda(ctx, id).
		Execute()
	if err != nil {
		return fmt.Errorf("error when calling `LambdaApi.DestroyLambda``: %v", err)
	}

	res, err := pollTask(ctx, taskResp.GetTask())
	if err != nil {
		return err
	}

	if res.GetStatus() == "FAILED" {
		return fmt.Errorf("error when destroying lambda: %v", res.GetDetails()["error"])
	}

	return nil
}

func DeployLambda(ctx context.Context, input CreateLambdaM, path string) (*api.Lambda, error) {
	lambda, err := CreateLambda(ctx, input, path)
	if err != nil {
		return nil, err
	}

	return StartLambda(ctx, lambda.GetId())
}
