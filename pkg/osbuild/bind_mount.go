package osbuild

type BindMountOptions struct {
	Source string `json:"source,omitempty"`
}

func (BindMountOptions) isMountOptions() {}

func NewBindMount(name, source, target string) *Mount {
	// XXX: validate source,target?

	return &Mount{
		Type:   "org.osbuild.bind",
		Name:   name,
		Target: target,
		Options: &BindMountOptions{
			Source: source,
		},
	}
}
