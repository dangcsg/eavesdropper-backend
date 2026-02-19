package handlers

import (
	apiErr "eavesdropper/api/error"
	"eavesdropper/api/middlewares"
	"eavesdropper/errs"
	"eavesdropper/services/audio"
	"eavesdropper/services/auth"
	"eavesdropper/services/transcribe"
	"eavesdropper/services/transcripts"
	"eavesdropper/services/users"
	"eavesdropper/services/whitelist"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Handles a transcription request.
// Before this is called, the client stores the audio file and a manifest describing it in the cloud storage.
// Downloads the manifest (sessionId query param identifies it), stores the audio locally.
// Checks if the user is allowed to transcribe it. (free audio seconds + subscription's are enough to cover the full audio)
// Stores the transcript, udates the user data. (deducts free credits fully, the rest of the audio is accounted as paid usage)
// Responds full transcript + id
func Transcribe(w http.ResponseWriter, r *http.Request) {

	fmt.Println("on transcribe handler")

	ctx := r.Context()

	token, ok := r.Context().Value(middlewares.AuthTokenKey).(string)
	if !ok {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to find auth token request context.")
		return
	}
	userId, err := auth.GetUserID(ctx, token)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusUnauthorized, "", "Token does not match user: "+err.Error())
		return
	}

	sessionId := r.URL.Query().Get("sessionId")
	if sessionId == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing sessionId")
		return
	}

	tmpDir, err := os.MkdirTemp("", "finalize-*")
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to create temporary directory to store audio files")
		return
	}
	defer os.RemoveAll(tmpDir)

	audioPath, _, audioSeconds, err := audio.ProcessAudioChunks(ctx, sessionId, userId, tmpDir)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to process audio chunks: "+err.Error())
		return
	}

	_, consumedFreeAudioSeconds, err := transcribe.TranscriptionAllowed(ctx, userId, audioSeconds)
	if err != nil && err != errs.ErrExceededSubscriptionTranscriptionLimits {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to check if the transcription is allowed: "+err.Error())
		return
	}
	if err == errs.ErrExceededSubscriptionTranscriptionLimits {
		apiErr.WriteJSONError(
			w,
			http.StatusForbidden,
			errs.ErrExceededSubscriptionTranscriptionLimits.Error(),
			"Subscription limits have been reached or will be with this transcription",
		)
		return
	}

	transcriptionResponse, err := transcribe.TranscribeAudioFile(ctx, audioPath, "audio/wav")
	if err != nil {
		apiErr.WriteJSONError(
			w,
			http.StatusInternalServerError,
			"",
			"Failed to transcribe audio file: "+err.Error(),
		)
		return
	}
	fmt.Println("Got gemini response text: ")
	if transcriptionResponse != nil {
		fmt.Println(transcriptionResponse.Text())
	} else {
		fmt.Println("nil gemini response")
	}

	savedTranscript, err := transcripts.SaveTranscript(
		ctx,
		transcriptionResponse,
		userId, sessionId,
		audioSeconds, consumedFreeAudioSeconds,
	)
	if err != nil {
		apiErr.WriteJSONError(
			w,
			http.StatusInternalServerError,
			"",
			"Failed to save transcript: "+err.Error(),
		)
		return
	}

	go users.DecrementFreeTier(userId, consumedFreeAudioSeconds)

	response := transcriptToResponse(savedTranscript)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func GetTranscript(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	transcriptID := r.PathValue("id")
	if transcriptID == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing transcript id")
		return
	}

	userId := ""
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		_userId, err := auth.GetUserID(ctx, token)
		if err == nil {
			userId = _userId
		}
	}

	transcript, transcriptOwnerID, err := transcripts.GetTranscriptByID(ctx, transcriptID)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusNotFound, "", "Transcript not found: "+err.Error())
		return
	}

	if !transcript.IsPrivate {
		response := transcriptToResponse(transcript)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	if userId == "" {
		apiErr.WriteJSONError(w, http.StatusUnauthorized, "", "Authentication required to access private transcript")
		return
	}

	if userId == transcriptOwnerID {
		response := transcriptToResponse(transcript)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	isWhitelisted, err := whitelist.IsUserWhitelistedForTranscript(ctx, transcriptOwnerID, transcriptID, userId)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to check whitelist: "+err.Error())
		return
	}

	if !isWhitelisted {
		apiErr.WriteJSONError(w, http.StatusForbidden, "", "Access denied")
		return
	}

	response := transcriptToResponse(transcript)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func UpdateTranscriptVisibility(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.PathValue("id")
	if userID == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing user id")
		return
	}

	transcriptId := r.PathValue("tId")
	if transcriptId == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing transcript id")
		return
	}

	visibleStr := r.URL.Query().Get("visible")
	if visibleStr == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Missing the 'visible' query param")
		return
	}
	visible := visibleStr == "true"

	err := transcripts.UpdateTranscriptPrivacy(ctx, userID, transcriptId, visible)
	if err != nil {
		apiErr.WriteJSONError(
			w,
			http.StatusInternalServerError,
			"",
			"Failed to make transcript public: "+err.Error(),
		)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpdateTranscriptTitle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.PathValue("id")
	if userID == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing user id")
		return
	}

	transcriptId := r.PathValue("tId")
	if transcriptId == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "missing transcript id")
		return
	}

	title := r.URL.Query().Get("tittle")
	if title == "" {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Missing the 'title' query param")
		return
	}

	err := transcripts.UpdateTranscriptTitle(ctx, userID, transcriptId, title)
	if err != nil {
		apiErr.WriteJSONError(
			w,
			http.StatusInternalServerError,
			"",
			"Failed to update transcript title: "+err.Error(),
		)
		return
	}

	w.WriteHeader(http.StatusOK)
}
