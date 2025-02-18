package experimental

import (
	"os"
	"strings"
)

const envKEY = "IMAGES_EXPERIMENTAL"

func Buildroot() string {
	env := os.Getenv(envKEY)
	if env == "" {
		return ""
	}
	for _, s := range strings.Split(env, ",") {
		if strings.HasPrefix(s, "force-buildroot=") {
			return strings.SplitN(s, "=", 2)[1]
		}
	}
	return ""
}
