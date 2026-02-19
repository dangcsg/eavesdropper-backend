package handlers

import (
	"context"
	apiErr "eavesdropper/api/error"
	"eavesdropper/api/middlewares"
	"eavesdropper/dtos/requests"
	"eavesdropper/dtos/responses"
	"eavesdropper/services/auth"
	"eavesdropper/services/transcripts"
	"eavesdropper/services/users"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func AddUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Printf("on add user handler")

	var newUserReq requests.NewUser
	if err := json.NewDecoder(r.Body).Decode(&newUserReq); err != nil {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Invalid request body: "+err.Error())
		return
	}

	token, ok := r.Context().Value(middlewares.AuthTokenKey).(string)
	if !ok {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to find auth token request context.")
		return
	}

	log.Printf("Got token %s", token)

	userID, err := auth.GetUserID(ctx, token)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusUnauthorized, "", "Failed to get userID from token: "+err.Error())
		return
	}

	log.Printf("Got userID %s", userID)

	if newUserReq.Email == "" {
		userRecord, err := auth.GetUserRecord(ctx, token)
		if err != nil {
			apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to get user record: "+err.Error())
			return
		}
		if userRecord.Email == "" {
			apiErr.WriteJSONError(
				w,
				http.StatusBadRequest,
				"",
				"An email must be provided when not using google oauth. "+
					"There was no email found in the user record",
			)
		}
		newUserReq.Email = userRecord.Email
	}

	if err := users.AddUser(userID, &newUserReq); err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to create user: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")

	var updateUserReq requests.UpdateUser
	if err := json.NewDecoder(r.Body).Decode(&updateUserReq); err != nil {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Invalid request body: "+err.Error())
		return
	}

	if err := users.UpdateUser(userId, &updateUserReq); err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to update user: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")

	user, err := users.GetUser(userId)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to get user: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func HandleAvailable(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	handle := r.URL.Query().Get("handle")
	if handle == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Missing the handle query param")
		return
	}

	available, err := users.HandleIsAvailable(ctx, handle)
	if err != nil {
		apiErr.WriteJSONError(
			w,
			http.StatusInternalServerError,
			"",
			"Failed to check if handle is in use: "+err.Error(),
		)
		return
	}

	response := map[string]interface{}{
		"available": available,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func GetUserTranscripts(w http.ResponseWriter, r *http.Request) {
	pageI, _ := parseQueryParamToInt(w, r, "pageI", false)
	pageSize, _ := parseQueryParamToInt(w, r, "pageSize", false)
	userId := r.PathValue("id")

	fmt.Printf("\nOn get user transcipts for page %v with page size %v", pageI, pageSize)

	transcripts, err := transcripts.GetUserTranscripts(r.Context(), userId, pageI, pageSize)
	if err != nil {
		apiErr.WriteJSONError(
			w,
			http.StatusInternalServerError,
			"",
			"Failed to get user transcripts: "+err.Error(),
		)
		return
	}

	// Convert resource models to response models
	responseTranscripts := make([]responses.TranscriptionResponse, len(transcripts))
	for i, transcript := range transcripts {
		responseTranscripts[i] = transcriptToResponse(&transcript)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseTranscripts)
}

func GetUserTranscript(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId := r.PathValue("id")
	if userId == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing user id")
		return
	}

	transcriptId := r.PathValue("tId")
	if transcriptId == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing transcript id")
		return
	}

	transcript, err := transcripts.GetUserTranscript(ctx, userId, transcriptId)
	if err != nil {
		apiErr.WriteJSONError(
			w,
			http.StatusInternalServerError,
			"",
			"Failed to get transcript: "+err.Error(),
		)
		return
	}

	response := transcriptToResponse(transcript)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func GetUserBillingCycleUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId := r.PathValue("id")
	if userId == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing user id")
		return
	}

	usage, err := users.GetCurrentBillingCycleUsage(ctx, userId)
	if err != nil {
		apiErr.WriteJSONError(
			w,
			http.StatusInternalServerError,
			"",
			"Failed to get billing cycle usage: "+err.Error(),
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(usage)
}
