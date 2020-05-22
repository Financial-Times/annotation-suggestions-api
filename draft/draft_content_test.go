package draft

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/Financial-Times/draft-content-suggestions/mocks"
	"github.com/stretchr/testify/assert"
)

func TestDraftContentAPI_IsGTGSuccess(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient)

	assert.NoError(t, err)

	msg, err := contentAPI.IsGTG(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "draft-content-public-read is healthy", msg)
}

func TestDraftContentAPI_IsGTGFailure503(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(false)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient)

	assert.NoError(t, err)

	_, err = contentAPI.IsGTG(context.Background())
	assert.Error(t, err)
}

func TestDraftContentAPI_IsGTGFailureInvalidEndpoint(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(false)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", ":#", http.DefaultClient, http.DefaultClient)

	assert.NoError(t, err)

	_, err = contentAPI.IsGTG(context.Background())

	var urlErr *url.Error
	if assert.Error(t, err) && errors.As(err, &urlErr) {
		assert.Equal(t, "parse", urlErr.Op)
	}
}

func TestDraftContentAPI_IsGTGFailureRequestError(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(false)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", "__gtg", http.DefaultClient, http.DefaultClient)

	assert.NoError(t, err)

	_, err = contentAPI.IsGTG(context.Background())
	var urlErr *url.Error
	if assert.Error(t, err) && errors.As(err, &urlErr) {
		assert.Equal(t, "Get", urlErr.Op)
	}
}

func TestDraftContentAPI_FetchDraftContentSuccess(t *testing.T) {
	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.ValidMockContentUUID)

	assert.NoError(t, err)
	assert.True(t, content != nil)
	assert.True(t, len(content) > 0)
}

func TestDraftContentAPI_FetchDraftContentMissing(t *testing.T) {

	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.MissingMockContentUUID)

	assert.NoError(t, err)
	assert.True(t, content == nil)
}

func TestDraftContentAPI_FetchDraftContentFailure(t *testing.T) {

	contentAPI, err := NewContentAPI("http://localhost/drafts/content", "http://localhost/__gtg", http.DefaultClient, http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.ValidMockContentUUID)

	assert.Error(t, err)
	assert.True(t, content == nil)
}

func TestDraftContentAPI_FetchDraftContentUnmappable(t *testing.T) {

	testServer := mocks.NewDraftContentTestServer(false)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.UnprocessableContentUUID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrDraftNotMappable))
	assert.True(t, content == nil)

}

func TestDraftContentAPI_FetchDraftContentNon200(t *testing.T) {

	testServer := mocks.NewDraftContentTestServer(true)
	defer testServer.Close()

	contentAPI, err := NewContentAPI(testServer.URL+"/drafts/content", testServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient)
	assert.NoError(t, err)

	content, err := contentAPI.FetchDraftContent(context.Background(), mocks.FailsRetrivalContentUuid)

	assert.Error(t, err)
	assert.EqualError(t, err, "error in draft content retrival status=500")
	assert.True(t, content == nil)
}
