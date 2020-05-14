package draft

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Financial-Times/draft-content-suggestions/commons"
	log "github.com/sirupsen/logrus"
)

var (
	ErrDraftNotMappable = errors.New("draft content is invalid for mapping")
)

func NewContentAPI(endpoint string, healthEndpoint string, httpClient *http.Client, healthHTTPClient *http.Client) (contentAPI ContentAPI, err error) {
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	contentAPI = &draftContentAPI{
		endpoint,
		healthEndpoint,
		httpClient,
		healthHTTPClient,
	}

	err = contentAPI.IsValid()
	if err != nil {
		return nil, err
	}

	return contentAPI, nil

}

// ContentApi for accessing to draft-content-api endpoint
type ContentAPI interface {
	FetchDraftContent(ctx context.Context, uuid string) (content []byte, err error)
	commons.Endpoint
}

type draftContentAPI struct {
	endpoint         string
	healthEndpoint   string
	httpClient       *http.Client
	healthHTTPClient *http.Client
}

func (d *draftContentAPI) FetchDraftContent(ctx context.Context, uuid string) ([]byte, error) {
	requestPath := d.endpoint + uuid
	request, err := http.NewRequest(http.MethodGet, requestPath, nil)
	if err != nil {
		return nil, err
	}

	response, err := d.httpClient.Do(request.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if response.StatusCode == http.StatusUnprocessableEntity {
		return nil, ErrDraftNotMappable
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error in draft content retrival status=%v", response.StatusCode)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (d *draftContentAPI) Endpoint() string {
	return d.endpoint
}

func (d *draftContentAPI) IsGTG(ctx context.Context) (string, error) {
	req, err := http.NewRequest(http.MethodGet, d.healthEndpoint, nil)
	if err != nil {
		log.WithError(err).WithField("healthEndpoint", d.healthEndpoint).Error("Error in creating GTG request to draft-content-public-read")
		return "", err
	}

	// We don't want logging for GTG requests in the middleware
	response, err := d.healthHTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		log.WithError(err).WithField("healthEndpoint", d.healthEndpoint).Error("Error in GTG request to draft-content-public-read")
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.WithField("healthEndpoint", d.healthEndpoint).
			WithField("status", response.StatusCode).
			Error("GTG for draft-content-public-read returned a non-200 HTTP status")
		return "", fmt.Errorf("GTG for draft-content-public-read returned a non-200 HTTP status: %v", response.StatusCode)
	}

	return "draft-content-public-read is healthy", nil
}

func (d *draftContentAPI) IsValid() error {
	return commons.ValidateEndpoint(d.endpoint)
}
