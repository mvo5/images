package imagefilter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gobwas/glob"

	"github.com/osbuild/images/pkg/distro"
)

func splitPrefixSearchTerm(s string) (string, string) {
	l := strings.SplitN(s, ":", 2)
	if len(l) == 1 {
		return "", l[0]
	}
	return l[0], l[1]
}

// newFilter creates an image filter based on the given filter terms. Glob like
// patterns (?, *) are supported, see fnmatch(3).
//
// Without a prefix in the filter term a simple name filtering is performed.
// With a prefix the specified property is filtered, e.g. "arch:i386". Adding
// filtering will narrow down the filtering (terms are combined via AND).
//
// The following prefixes are supported:
// "distro:" - the distro name, e.g. rhel-9, or fedora*
// "arch:" - the architecture, e.g. x86_64
// "type": - the image type, e.g. ami, or qcow?
// "bootmode": - the bootmode, e.g. "legacy", "uefi", "hybrid"
func newFilter(sl ...string) (*filter, error) {
	filter := &filter{
		prefix:  make([]string, len(sl)),
		pattern: make([]glob.Glob, len(sl)),
	}
	for i, s := range sl {
		prefix, searchTerm := splitPrefixSearchTerm(s)
		if !slices.Contains(supportedFilters, prefix) {
			return nil, fmt.Errorf("unsupported filter prefix: %q", prefix)
		}
		gl, err := glob.Compile(searchTerm)
		if err != nil {
			return nil, err
		}
		filter.prefix[i] = prefix
		filter.pattern[i] = gl
	}
	return filter, nil
}

var supportedFilters = []string{
	"", "distro", "arch", "type", "bootmode",
}

// filter provides a way to filter a list of image defintions for the
// given filter terms.
type filter struct {
	prefix  []string
	pattern []glob.Glob
}

// Matches returns true if the given (distro,arch,imgType) tuple matches
// the filter expressions
func (fl filter) Matches(distro distro.Distro, arch distro.Arch, imgType distro.ImageType) bool {
	m := true
	for i, gl := range fl.pattern {
		switch fl.prefix[i] {
		case "":
			// no prefix, do a "fuzzy" search accross the common
			// things users may want
			m1 := gl.Match(distro.Name())
			m2 := gl.Match(arch.Name())
			m3 := gl.Match(imgType.Name())
			m = m && (m1 || m2 || m3)
		case "distro":
			m = m && gl.Match(distro.Name())
		case "arch":
			m = m && gl.Match(arch.Name())
		case "type":
			m = m && gl.Match(imgType.Name())
			// mostly here to show how flexible this is
		case "bootmode":
			m = m && gl.Match(imgType.BootMode().String())
		}
	}
	return m
}
