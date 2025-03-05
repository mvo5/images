package inputs

import (
	"github.com/osbuild/images/pkg/rpmmd"
)

// RepoContainerConfig contains the repository and container registry
// config used to instanciate the given image
type RepoContainerConfig struct {
	Repos []rpmmd.RepoConfig

	// Container reference for the (optional) bootstrap container
	// used when doing a cross-arch or cross-distro bootstrap of
	// the buildroot
	BootstrapContainerRef string
}
