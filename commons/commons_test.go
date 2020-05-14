package commons

import (
	"fmt"
	"net/http"
	"testing"

	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

const (
	validEndpoint   = "http://some.valid.url/with/sub/resources:8080"
	invalidEndpoint = "/missing.com/scheme/or/uri/type"
	invalidUUID     = "random:invalid-string"
)

func TestValidateEndpointSuccess(t *testing.T) {
	assert.NoError(t, ValidateEndpoint(validEndpoint))
}

func TestValidateEndpointFailure(t *testing.T) {
	isValid := ValidateEndpoint(invalidEndpoint)
	assert.Error(t, isValid)
}

func TestValidateUUIDSuccess(t *testing.T) {
	v4, err := uuid.NewV4()
	assert.NoError(t, err)
	assert.NoError(t, ValidateUUID(v4.String()))
}

func TestValidateUUIDFailure(t *testing.T) {
	isValid := ValidateUUID(invalidUUID)

	if isValid == nil {
		assert.Fail(t, fmt.Sprintf("UUID validation should've failed for: %v", invalidUUID))
	}
}

func TestNewContextFromRequest(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, validEndpoint, nil)
	contextFromRequest := NewContextFromRequest(request)

	trxID, ok := contextFromRequest.Value(transactionidutils.TransactionIDKey).(string)

	assert.True(t, ok)
	assert.NotEmpty(t, trxID)
}
