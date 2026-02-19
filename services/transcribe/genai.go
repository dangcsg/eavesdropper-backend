package transcribe

import (
	"context"
	cnfg "eavesdropper/configurations"

	"google.golang.org/genai"
)

const projectID = "eavesdropper-4f10b"
const location = "europe-west4"

func getGenaiClient(ctx context.Context) (*genai.Client, error) {

	if cnfg.SelectedDeployment == cnfg.Cloud {
		// THIS SDK is not supported for file uploads. The other works fine in production
		// fmt.Println("Initializing genAI client via vertexBackend")
		// return genai.NewClient(ctx, &genai.ClientConfig{
		// 	Project:  projectID,
		// 	Location: location,
		// 	Backend:  genai.BackendVertexAI,
		// })
	}

	return genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cnfg.GenAIApiKey,
		Backend: genai.BackendGeminiAPI,
	})
}

type GeminiModel string

// todo Read this to check if the model in the prompt req matches one of these, to not allow users to request with advanced unsuported models (should they manually modify the request)
const (
	GeminiFlash20 GeminiModel = "gemini-2.0-flash"
	GeminiFlash25 GeminiModel = "gemini-2.5-flash"
	// Deprecated, dont work in prod
	// GeminiFlash   GeminiModel = "gemini-1.5-flash"
	// GeminiFlash8B GeminiModel = "gemini-1.5-flash-8b"
	// GeminiPro     GeminiModel = "gemini-1.5-pro"
)
