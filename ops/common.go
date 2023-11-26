package ops

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	api "github.com/onpremless/go-client"
)

var client *api.APIClient

func init() {
	config := api.NewConfiguration()
	config.Servers = api.ServerConfigurations{
		{
			// XXX: WOAH JASON, NICE HARD-CODED URL!
			URL: "http://localhost:8081",
		},
	}

	client = api.NewAPIClient(config)
}

func upload(ctx context.Context, path string, isDir bool) (string, error) {
	if isDir {
		var err error
		path, err = tarPath(path)
		if err != nil {
			return "", err
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	defer func() {
		if isDir {
			os.Remove(path)
		}
	}()

	uploadResp, _, err := client.UploadAPI.Upload(ctx).File(file).Execute()
	if err != nil {
		return "", fmt.Errorf("error when calling `UploadApi.Upload``: %v", err)
	}

	return uploadResp.GetId(), nil
}

func pollTask(ctx context.Context, id string) (*api.TaskStatus, error) {
	for {
		resp, _, err := client.TaskAPI.
			GetTask(ctx, id).
			Execute()
		if err != nil {
			return nil, fmt.Errorf("error when calling `TaskApi.GetTask``: %v", err)
		}

		if resp.GetStatus() != "PENDING" {
			return resp, nil
		}

		time.Sleep(time.Second)
	}
}

func tarPath(src string) (string, error) {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", src)
	}

	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)

	err := filepath.Walk(src, func(file string, info os.FileInfo, lerr error) error {
		header, err := tar.FileInfoHeader(info, file)
		if err != nil {
			return err
		}

		header.Name = filepath.Base(file)

		if err := writer.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(writer, data); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	file, err := os.CreateTemp("", "lambda-")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, &buffer); err != nil {
		os.Remove(file.Name())
		return "", err
	}

	return file.Name(), nil
}
