package utils

import (
	"context"
	"encoding/json"
	"net/http"
)

type GetRequest struct {
	Url     string
	Headers map[string]string
}

func HTTPGet[T any](ctx context.Context, req *GetRequest) (*T, error) {
	request, err := http.NewRequest("GET", req.Url, nil)
	if err != nil {
		return nil, err
	}
	if len(req.Headers) > 0 {
		for k, v := range req.Headers {
			request.Header.Add(k, v)
		}
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var data T
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
