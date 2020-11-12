package client

import (
	"fmt"
	"net/http"
)

type tokenClient struct {
	token      string
	httpClient *http.Client
}

func (t tokenClient) BaseURL() string {
	return DefaultBaseURL
}

func (t tokenClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Token %v", t.token))

	return handleRequest(t.httpClient, req)
}
