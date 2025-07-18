package defs

import (
	"fmt"
	"regexp"

	"github.com/osbuild/images/pkg/distro"
)

// ParseID parse the given nameVer into a distro.ID. It will also also
// search through all distros for a name match. This is needed to
// support distro names like "rhel-810" without dots.
//
// If no name is found it will return nil, nil
func ParseID(nameVer string) (*distro.ID, error) {
	_, id, err := distroAndIDFor(nameVer)
	if err != nil {
		return nil, err
	}
	if id != nil {
		return id, nil
	}
	return nil, nil
}

func distroAndIDFor(nameVer string) (*DistroYAML, *distro.ID, error) {
	distros, err := loadDistros()
	if err != nil {
		return nil, nil, err
	}

	for _, d := range distros.Distros {
		if d.InternalName == nameVer {
			id, err := distro.ParseID(nameVer)
			return &d, id, err
		}

		re, err := regexp.Compile(d.InternalName)
		if err != nil {
			return nil, nil, err
		}
		if l := re.FindStringSubmatch(nameVer); len(l) > 2 {
			switch len(l) {
			case 3:
				id, err := distro.ParseID(fmt.Sprintf("%s-%s", l[re.SubexpIndex("name")], l[re.SubexpIndex("major")]))
				return &d, id, err
			case 4:
				id, err := distro.ParseID(fmt.Sprintf("%s-%s.%s", l[re.SubexpIndex("name")], l[re.SubexpIndex("major")], l[re.SubexpIndex("minor")]))
				return &d, id, err
			default:
				panic(fmt.Errorf("internal error: %v submatches for %v", len(l), nameVer))
			}
		}
	}
	return nil, nil, nil
}
