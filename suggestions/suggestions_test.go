package suggestions

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/Financial-Times/draft-content-suggestions/mocks"
	log "github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

const apiKey = "12345"

func TestUmbrellaAPI_IsGTGSuccess(t *testing.T) {
	hook := logTest.NewGlobal()
	testServer := mocks.NewUmbrellaTestServer(true)
	defer testServer.Close()

	umbrellaAPI, err := NewUmbrellaAPI(testServer.URL+"/content/suggest", testServer.URL+"/content/suggest/__gtg", apiKey, http.DefaultClient, http.DefaultClient)

	assert.NoError(t, err)

	msg, err := umbrellaAPI.IsGTG(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "UPP suggestions API is healthy", msg)
	assert.Empty(t, hook.Entries)
}

func TestUmbrellaAPI_IsGTGFailure503(t *testing.T) {
	hook := logTest.NewGlobal()
	testServer := mocks.NewUmbrellaTestServer(false)
	defer testServer.Close()

	umbrellaAPI, err := NewUmbrellaAPI(testServer.URL+"/content/suggest", testServer.URL+"/content/suggest/__gtg", apiKey, http.DefaultClient, http.DefaultClient)

	assert.NoError(t, err)

	_, err = umbrellaAPI.IsGTG(context.Background())
	assert.Error(t, err)
	assert.Error(t, err)
	assert.Len(t, hook.AllEntries(), 1)
	assert.Equal(t, log.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(t, "GTG for UPP suggestions API returned a non-200 HTTP status", hook.LastEntry().Message)
	assert.Equal(t, http.StatusServiceUnavailable, hook.LastEntry().Data["status"])
	assert.Equal(t, testServer.URL+"/content/suggest/__gtg", hook.LastEntry().Data["healthEndpoint"])
}

func TestUmbrellaAPI_IsGTGFailureInvalidEndpoint(t *testing.T) {
	hook := logTest.NewGlobal()
	testServer := mocks.NewUmbrellaTestServer(false)
	defer testServer.Close()

	umbrellaAPI, err := NewUmbrellaAPI(testServer.URL+"/content/suggest", ":#", apiKey, http.DefaultClient, http.DefaultClient)

	assert.NoError(t, err)

	_, err = umbrellaAPI.IsGTG(context.Background())

	var urlErr *url.Error
	if assert.Error(t, err) && errors.As(err, &urlErr) {
		assert.Equal(t, "parse", urlErr.Op)
	}

	assert.Len(t, hook.AllEntries(), 1)
	assert.Equal(t, log.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(t, "Error in creating GTG request to UPP suggestions API", hook.LastEntry().Message)
	assert.Equal(t, ":#", hook.LastEntry().Data["healthEndpoint"])
}

func TestUmbrellaAPI_IsGTGFailureRequestError(t *testing.T) {
	hook := logTest.NewGlobal()
	testServer := mocks.NewUmbrellaTestServer(false)
	defer testServer.Close()

	umbrellaAPI, err := NewUmbrellaAPI(testServer.URL+"/content/suggest", "__gtg", apiKey, http.DefaultClient, http.DefaultClient)

	assert.NoError(t, err)

	_, err = umbrellaAPI.IsGTG(context.Background())

	var urlErr *url.Error
	if assert.Error(t, err) && errors.As(err, &urlErr) {
		assert.Equal(t, "Get", urlErr.Op)
	}

	assert.Len(t, hook.AllEntries(), 1)
	assert.Equal(t, log.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(t, "Error in GTG request to UPP suggestions API", hook.LastEntry().Message)
	assert.Equal(t, "__gtg", hook.LastEntry().Data["healthEndpoint"])
}

func TestUmbrellaAPI_FetchSuggestions(t *testing.T) {

	mockDraftContent := newMockDraftContent()

	testServer := mocks.NewUmbrellaTestServer(true)
	defer testServer.Close()

	umbrellaAPI, err := NewUmbrellaAPI(testServer.URL+"/content/suggest", testServer.URL+"/content/suggest/__gtg", apiKey, http.DefaultClient, http.DefaultClient)
	assert.NoError(t, err)

	suggestions, err := umbrellaAPI.FetchSuggestions(context.Background(), mockDraftContent)

	assert.NoError(t, err)
	assert.True(t, suggestions != nil)
}
func TestUmbrellaAPI_FetchDraftContentFailure(t *testing.T) {

	testServer := mocks.NewUmbrellaTestServer(false)
	defer testServer.Close()

	contentAPI, err := NewUmbrellaAPI(testServer.URL+"/content/suggest", testServer.URL+"/content/suggest/__gtg", apiKey, http.DefaultClient, http.DefaultClient)
	assert.NoError(t, err)

	suggestions, err := contentAPI.FetchSuggestions(context.Background(), newMockDraftContent())

	assert.Error(t, err)
	assert.True(t, suggestions == nil)
}

func newMockDraftContent() []byte {

	mockContent := make(map[string]interface{})

	mockContent["uuid"] = "9d5e441e-0b02-11e8-8eb7-42f857ea9f0"
	mockContent["body"] = "<body><content data-embedded=\"true\" id=\"c0cc4ca2-0b43-11e8-24ad-bec2279df517\" type=\"http://www.ft.com/ontology/content/ImageSet\"></content><p>US stocks see-sawed in early trading on Tuesday, as volatility on global markets intensified, breaking an extended period of calm for investors.xxxx</body>"
	mockContent["title"] = "Wall Street stocks xxx"
	mockContent["byline"] = "Eric Platt in New York, Michael Hunter and Adam Samson in London"

	bytes, _ := json.Marshal(mockContent)

	return bytes
}
