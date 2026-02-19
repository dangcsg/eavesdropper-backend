package configurations

//region Change this to control the enviroment
var SelectedBackendMode = Development
var SelectedDeployment = Localhost

//endregion

type BackendMode string

const (
	Development BackendMode = "development"
	Production  BackendMode = "production"
)

type Deployment string

const (
	Localhost Deployment = "localhost"
	Cloud     Deployment = "cloud"
)

// Audio formats supported
type supportedAudioFormat string

const (
	Wav supportedAudioFormat = "wav"
	// todo add others
)

var supportedAudioFormats = map[supportedAudioFormat]bool{
	Wav: true,
}

func IsSupportedAduioFormat(format string) bool {
	_, ok := supportedAudioFormats[supportedAudioFormat(format)]
	return ok
}
