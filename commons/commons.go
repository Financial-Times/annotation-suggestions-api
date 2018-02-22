package commons

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Financial-Times/go-ft-http-transport/transport"
	tidutils "github.com/Financial-Times/transactionid-utils-go"
	"github.com/satori/go.uuid"
	"time"
)

// Common type/behaviour definition for an endpoint
type Endpoint interface {
	// Endpoint
	// Returns the endpoint
	Endpoint() string

	// IsValid
	// Validates the structure of the url/uri(s)
	IsValid() error

	// IsGTG
	// Checks if this endpoint is actually reachable and performing as expected
	IsGTG(ctx context.Context) (string, error)
}

type message struct {
	Message string `json:"message"`
}

// WriteJSONMessage writes the msg provided as encoded json with the proper content type header added.
func WriteJSONMessage(w http.ResponseWriter, status int, msg string) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	return enc.Encode(&message{Message: msg})
}

// NewContextFromRequest provides a new context including a trxId
// from the request or if missing, a brand new trxId.
func NewContextFromRequest(r *http.Request) context.Context {
	return tidutils.TransactionAwareContext(context.Background(), tidutils.GetTransactionIDFromRequest(r))
}

// ValidateEndpoints provides url/uri level validation, it does not make any actual http(s) requests
func ValidateEndpoint(endpoint string) error {

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return errors.New(fmt.Sprintf("Missing scheme in endpoint: %v", endpoint))
	}
	_, err := url.ParseRequestURI(endpoint)

	if err != nil {
		return errors.New(fmt.Sprintln("Invalid endpoint configuration:", err, " for:", endpoint))
	}

	return nil
}

// ValidateUUID checks the uuid string for supported formats
func ValidateUUID(u string) error {
	_, err := uuid.FromString(u)
	return err
}

// NewFTHttpClient provides a FT compliant http client.
func NewFTHttpClient(platform string, systemCode string, timeout time.Duration) *http.Client {
	delegatingTransport := transport.NewTransport().WithStandardUserAgent(platform, systemCode)
	return &http.Client{Timeout: timeout, Transport: delegatingTransport}
}
