package handlers

import (
	apiErr "eavesdropper/api/error"
	"eavesdropper/dtos/resources"
	"eavesdropper/dtos/responses"
	"net/http"
	"strconv"
)

func parsePathValueToInt(w http.ResponseWriter, r *http.Request, key string, mandatory bool) (int, bool) {
	value := r.PathValue(key)
	if value == "" {
		if mandatory {
			apiErr.WriteJSONError(w, http.StatusBadRequest, "MISSING_PATH_PARAM", "Missing required path parameter: "+key)
		}
		return 0, false
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "INVALID_PATH_PARAM", "Invalid integer value for path parameter: "+key)
		return 0, false
	}

	return intValue, true
}

func parseQueryParamToInt(w http.ResponseWriter, r *http.Request, key string, mandatory bool) (int, bool) {
	value := r.URL.Query().Get(key)
	if value == "" {
		if mandatory {
			apiErr.WriteJSONError(w, http.StatusBadRequest, "MISSING_QUERY_PARAM", "Missing required query parameter: "+key)
		}
		return 0, false
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "INVALID_QUERY_PARAM", "Invalid integer value for query parameter: "+key)
		return 0, false
	}

	return intValue, true
}

// Helper function to convert Transcript resource to TranscriptionResponse
func transcriptToResponse(transcript *resources.Transcript) responses.TranscriptionResponse {
	return responses.TranscriptionResponse{
		ID:                        transcript.ID,
		RecordingSessionID:        transcript.RecordingSessionID,
		Tittle:                    transcript.Tittle,
		Content:                   transcript.Content,
		ConsumedInputAudioSeconds: transcript.ConsumedInputAudioSeconds,
		ConsumedFreeAudioSeconds:  transcript.ConsumedFreeAudioSeconds,
		CreatedAt:                 transcript.CreatedAt,
	}
}
