package handlers

import (
	"encoding/json"
	"net/http"

	apiErr "eavesdropper/api/error"
	"eavesdropper/dtos/requests"
	"eavesdropper/services/whitelist"
)

func AddUsersToTranscriptWhitelist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.PathValue("id")
	if userID == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing user id")
		return
	}

	transcriptID := r.PathValue("tId")
	if transcriptID == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing transcript id")
		return
	}

	var req requests.AddToWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Invalid request body: "+err.Error())
		return
	}

	if len(req.Users) == 0 {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "users list cannot be empty")
		return
	}

	err := whitelist.AddUsersToTranscriptWhitelist(ctx, userID, transcriptID, &req)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to add users to whitelist: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func RemoveUsersFromTranscriptWhitelist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.PathValue("id")
	if userID == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing user id")
		return
	}

	transcriptID := r.PathValue("tId")
	if transcriptID == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing transcript id")
		return
	}

	var req requests.RemoveFromWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Invalid request body: "+err.Error())
		return
	}

	if len(req.Handles) == 0 && len(req.Emails) == 0 {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "either handles or emails must be provided")
		return
	}

	err := whitelist.RemoveUsersFromTranscriptWhitelist(ctx, userID, transcriptID, &req)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to remove users from whitelist: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetTranscriptWhitelist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.PathValue("id")
	if userID == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing user id")
		return
	}

	transcriptID := r.PathValue("tId")
	if transcriptID == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing transcript id")
		return
	}

	whitelistResponse, err := whitelist.GetTranscriptWhitelist(ctx, userID, transcriptID)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to get transcript whitelist: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(whitelistResponse)
}
