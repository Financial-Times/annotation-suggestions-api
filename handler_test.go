package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/mocks"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestRequestHandlerSuccess(t *testing.T) {
	resp, err := handleTestRequest("/drafts/content/" + mocks.ValidMockContentUUID + "/suggestions")

	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestHandlerContentNotFound(t *testing.T) {
	resp, err := handleTestRequest("/drafts/content/" + mocks.MissingMockContentUUID + "/suggestions")

	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRequestHandlerContentNotMappable(t *testing.T) {

	resp, err := handleTestRequest("/drafts/content/" + mocks.UnprocessableContentUUID + "/suggestions")

	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestRequestHandlerContentInvalidUUID(t *testing.T) {

	resp, err := handleTestRequest("/drafts/content/invaliduuid/suggestions")

	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func handleTestRequest(urlpath string) (resp *http.Response, err error) {
	draftContentTestServer := mocks.NewDraftContentTestServer(true)
	umbrellaTestServer := mocks.NewUmbrellaTestServer(true)

	defer draftContentTestServer.Close()
	defer umbrellaTestServer.Close()

	contentAPI, _ := draft.NewContentAPI(draftContentTestServer.URL+"/drafts/content", draftContentTestServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient)
	umbrellaAPI, _ := suggestions.NewUmbrellaAPI(umbrellaTestServer.URL, umbrellaTestServer.URL+"/__gtg", "12345", http.DefaultClient, http.DefaultClient)

	requestHandler := requestHandler{contentAPI, umbrellaAPI}

	r := mux.NewRouter()
	r.HandleFunc("/drafts/content/{uuid}/suggestions", requestHandler.draftContentSuggestionsRequest)
	ts := httptest.NewServer(r)

	defer ts.Close()

	return http.Get(ts.URL + urlpath)

}
