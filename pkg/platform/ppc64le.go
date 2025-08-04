package platform

import (
	"github.com/osbuild/images/pkg/arches"
)

type PPC64LE struct {
	BasePlatform
	BIOS bool
}

func (p *PPC64LE) GetArch() arches.Arch {
	return arches.PPC64LE
}

func (p *PPC64LE) GetBIOSPlatform() string {
	if p.BIOS {
		return "powerpc-ieee1275"
	}
	return ""
}

func (p *PPC64LE) GetPackages() []string {
	packages := p.BasePlatform.FirmwarePackages

	if p.BIOS {
		packages = append(packages,
			"dracut-config-generic",
			"powerpc-utils",
			"grub2-ppc64le",
			"grub2-ppc64le-modules",
		)
	}

	return packages
}

func (p *PPC64LE) GetBuildPackages() []string {
	packages := []string{}

	if p.BIOS {
		packages = append(packages,
			"grub2-ppc64le",
			"grub2-ppc64le-modules",
		)
	}

	return packages
}

func (p *PPC64LE) GetBootloader() Bootloader {
	return BOOTLOADER_GRUB2
}
