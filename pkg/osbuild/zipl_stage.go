package osbuild

// The ZiplStageOptions describe how to create zipl stage
//
// The only configuration option available is a boot timeout and it is optional
type ZiplStageOptions struct {
	Timeout int `json:"timeout,omitempty"`
}

type ZiplStage struct {
	Type    string            `json:"type"`
	Options *ZiplStageOptions `json:"options,omitempty"`
}

// NewZiplStage creates a new zipl Stage object.
func NewZiplStage(options *ZiplStageOptions) Stage {
	return &ZiplStage{
		Type:    "org.osbuild.zipl",
		Options: options,
	}
}

func (st *ZiplStage) MarshalJSON() ([]byte, error) {
	return json.Marshal(st)
}
