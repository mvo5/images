package osbuild

type BootupdStageOptionsBios struct {
	Device    string `json:"device"`
	Partition int    `json:"partition,omitempty"`
}

type BootupdStageOptions struct {
	Deployment    OSTreeDeployment        `json:"deployment"`
	StaticConfigs bool                    `json:"static-configs"`
	Bios          BootupdStageOptionsBios `json:"bios"`
}

func (BootupdStageOptions) isStageOptions() {}

func NewBootupdStage(opts *BootupdStageOptions, devices *Devices, mounts *Mounts) *Stage {
	return &Stage{
		Type:    "org.osbuild.bootupd",
		Options: opts,
		Devices: *devices,
		Mounts:  *mounts,
	}
}
