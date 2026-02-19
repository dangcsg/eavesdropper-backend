package configurations

import "os"

var GenAIApiKey = os.Getenv("GENAI_API_KEY")

var GeminiAudioInputSecondsToTokenRate = 32 // one second costs 32 tokens
var GeminiAudioInputMaxSeconds = 9.5 * 60 * 60
