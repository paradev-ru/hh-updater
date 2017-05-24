package hhclient

import (
	"fmt"
	"net/http"
)

type TokenTransport struct {
	AccessToken string
}

func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AccessToken))
	return http.DefaultTransport.RoundTrip(req)
}

func TokenHTTPClient(accessToken string) *http.Client {
	return &http.Client{
		Transport: &TokenTransport{
			AccessToken: accessToken,
		},
	}
}
