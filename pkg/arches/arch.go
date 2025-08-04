package arches

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/osbuild/images/internal/common"
)

type Arch uint64

const ( // architecture enum
	UNSET Arch = iota
	AARCH64
	PPC64LE
	S390X
	X86_64
	RISCV64
)

func (a Arch) String() string {
	switch a {
	case UNSET:
		return "unset"
	case AARCH64:
		return "aarch64"
	case PPC64LE:
		return "ppc64le"
	case S390X:
		return "s390x"
	case X86_64:
		return "x86_64"
	case RISCV64:
		return "riscv64"
	default:
		panic("invalid architecture")
	}
}

func (a *Arch) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a, err = FromString(s)
	return err
}

func (a *Arch) UnmarshalYAML(unmarshal func(any) error) error {
	return common.UnmarshalYAMLviaJSON(a, unmarshal)
}

func FromString(a string) (Arch, error) {
	switch a {
	case "amd64", "x86_64":
		return X86_64, nil
	case "arm64", "aarch64":
		return AARCH64, nil
	case "s390x":
		return S390X, nil
	case "ppc64le":
		return PPC64LE, nil
	case "riscv64":
		return RISCV64, nil
	default:
		return UNSET, fmt.Errorf("unsupported architecture %q", a)
	}
}

var runtimeGOARCH = runtime.GOARCH

func Current() Arch {
	return common.Must(FromString(runtimeGOARCH))
}

func IsX86_64() bool {
	return Current() == X86_64
}

func IsAarch64() bool {
	return Current() == AARCH64
}

func IsPPC() bool {
	return Current() == PPC64LE
}

func IsS390x() bool {
	return Current() == S390X
}

func IsRISCV64() bool {
	return Current() == RISCV64
}
