package main

import (
	"net/http"

	logger "github.com/Financial-Times/go-logger/v2"
	tidutils "github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/mux"

	"github.com/Financial-Times/draft-content-suggestions/commons"
	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
)

type requestHandler struct {
	dca draft.ContentAPI
	sua suggestions.UmbrellaAPI
	log *logger.UPPLogger
}

func (rh *requestHandler) draftContentSuggestionsRequest(writer http.ResponseWriter, request *http.Request) {
	uuid := mux.Vars(request)["uuid"]
	log := rh.log.WithField(tidutils.TransactionIDKey, tidutils.GetTransactionIDFromRequest(request)).WithField("uuid", uuid)

	err := commons.ValidateUUID(uuid)
	if err != nil {
		msg := "Invalid UUID"
		log.WithError(err).Warn(msg)
		_ = commons.WriteJSONMessage(writer, http.StatusBadRequest, msg)
		return
	}

	ctx := commons.NewContextFromRequest(request)
	content, err := rh.dca.FetchDraftContent(ctx, uuid)
	if err == draft.ErrDraftNotMappable {
		msg := "Could not provide suggestions for content, as we are unable to map it"
		log.WithError(err).Info(msg)
		_ = commons.WriteJSONMessage(writer, http.StatusUnprocessableEntity, msg)
		return
	}
	if err != nil {
		msg := "Draft content api retrieval has failed."
		log.WithError(err).Error(msg)
		_ = commons.WriteJSONMessage(writer, http.StatusInternalServerError, msg)
		return
	}
	if content == nil {
		msg := "No draft content for UUID"
		log.Warn(msg)
		_ = commons.WriteJSONMessage(writer, http.StatusNotFound, msg)
		return
	}

	suggestion, err := rh.sua.FetchSuggestions(ctx, content)
	if err != nil {
		msg := "Suggestions umbrella api access has failed"
		log.WithError(err).Error(msg)
		_ = commons.WriteJSONMessage(writer, http.StatusServiceUnavailable, msg)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(suggestion)
	if err != nil {
		// could be related to intermittent/temporary network issues
		// or original Tagme request is no more waiting for a response.
		log.WithError(err).Error("Failed responding to draft content suggestions request")
	}

}
