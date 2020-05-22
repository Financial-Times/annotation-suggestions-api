package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/mocks"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
)

func TestRequestHandlerSuccess(t *testing.T) {
	resp, err := handleTestRequest("/drafts/content/" + mocks.ValidMockContentUUID + "/suggestions")

	resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestHandlerContentNotFound(t *testing.T) {
	resp, err := handleTestRequest("/drafts/content/" + mocks.MissingMockContentUUID + "/suggestions")

	resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRequestHandlerContentNotMappable(t *testing.T) {

	resp, err := handleTestRequest("/drafts/content/" + mocks.UnprocessableContentUUID + "/suggestions")

	resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestRequestHandlerContentInvalidUUID(t *testing.T) {

	resp, err := handleTestRequest("/drafts/content/invaliduuid/suggestions")

	resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func handleTestRequest(urlpath string) (resp *http.Response, err error) {
	draftContentTestServer := mocks.NewDraftContentTestServer(true)
	umbrellaTestServer := mocks.NewUmbrellaTestServer(true)

	defer draftContentTestServer.Close()
	defer umbrellaTestServer.Close()

	log := logger.NewUPPLogger("Test", "PANIC")
	contentAPI, _ := draft.NewContentAPI(draftContentTestServer.URL+"/drafts/content", draftContentTestServer.URL+"/__gtg", http.DefaultClient, http.DefaultClient)
	umbrellaAPI, _ := suggestions.NewUmbrellaAPI(umbrellaTestServer.URL, umbrellaTestServer.URL+"/__gtg", "12345", http.DefaultClient, http.DefaultClient)

	rh := requestHandler{contentAPI, umbrellaAPI, log}

	r := mux.NewRouter()
	r.HandleFunc("/drafts/content/{uuid}/suggestions", rh.draftContentSuggestionsRequest)
	ts := httptest.NewServer(r)

	defer ts.Close()

	return http.Get(ts.URL + urlpath)

}
