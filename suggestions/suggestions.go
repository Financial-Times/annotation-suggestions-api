package suggestions

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Financial-Times/draft-content-suggestions/commons"
)

const APIKeyHeader = "X-Api-Key"

func NewUmbrellaAPI(endpoint string, gtgEndpoint string, apiKey string, httpClient *http.Client, healthHTTPClient *http.Client) (UmbrellaAPI, error) {
	umbrellaAPI := &umbrellaAPI{endpoint, gtgEndpoint, apiKey, httpClient, healthHTTPClient}

	err := umbrellaAPI.IsValid()
	if err != nil {
		return nil, err
	}

	return umbrellaAPI, nil
}

type UmbrellaAPI interface {
	// FetchSuggestions
	// Makes a API request to Suggestions Umbrella and returns the
	// []byte body
	FetchSuggestions(ctx context.Context, content []byte) (suggestion []byte, err error)

	// Embedded Endpoint interface, check its godoc
	commons.Endpoint
}

type umbrellaAPI struct {
	endpoint         string
	gtgEndpoint      string
	apiKey           string
	httpClient       *http.Client
	healthHTTPClient *http.Client
}

func (u *umbrellaAPI) FetchSuggestions(ctx context.Context, content []byte) (suggestion []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.endpoint, bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}

	req.Header.Set(APIKeyHeader, u.apiKey)

	res, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("suggestions Umbrella endpoint fail: %s", res.Status)
	}

	suggestion, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading the response body from Suggestions Umbrella endpoint: %w", err)
	}

	return suggestion, nil
}

func (u *umbrellaAPI) Endpoint() string {
	return u.endpoint
}

func (u *umbrellaAPI) IsGTG(ctx context.Context) (string, error) {
	req, err := http.NewRequest(http.MethodGet, u.gtgEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("error creating GTG request: %w", err)
	}

	req.Header.Set(APIKeyHeader, u.apiKey)

	response, err := u.healthHTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("error sending GTG request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 HTTP response: %d", response.StatusCode)
	}

	return "UPP suggestions API is healthy", nil
}

func (u *umbrellaAPI) IsValid() error {
	return commons.ValidateEndpoint(u.endpoint)
}
