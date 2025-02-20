package experimental

import (
	"os"
	"strings"
)

const envKEY = "IMAGES_EXPERIMENTAL"

func experimentalOptions() map[string]string {
	expMap := map[string]string{}

	env := os.Getenv(envKEY)
	if env == "" {
		return expMap
	}

	for _, s := range strings.Split(env, ",") {
		l := strings.SplitN(s, "=", 2)
		expMap[l[0]] = l[1]
	}
	return expMap
}

func Buildroot() string {
	expMap := experimentalOptions()
	return expMap["buildroot"]
}
