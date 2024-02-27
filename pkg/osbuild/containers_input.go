package osbuild

import (
	"github.com/osbuild/images/pkg/container"
)

type ContainersInputSourceRef struct {
	Name string `json:"name"`
}

type ContainersInput struct {
	inputCommon
	References map[string]ContainersInputSourceRef `json:"references"`
}

func (c ContainersInput) isStageInputs() {}

func newContainersInputForSources(containers []container.Spec, forLocal bool) ContainersInput {
	refs := make(map[string]ContainersInputSourceRef, len(containers))
	for _, c := range containers {
		if !forLocal && c.LocalStorage {
			continue
		}
		ref := ContainersInputSourceRef{
			Name: c.LocalName,
		}
		refs[c.ImageID] = ref
	}

	var typ string
	if forLocal {
		typ = "org.osbuild.containers-storage"
	} else {
		typ = "org.osbuild.containers"
	}

	return ContainersInput{
		References: refs,
		inputCommon: inputCommon{
			Type:   typ,
			Origin: InputOriginSource,
		},
	}
}

func NewContainersInputForSources(containers []container.Spec) ContainersInput {
	return newContainersInputForSources(containers, false)
}

func NewLocalContainersInputForSources(containers []container.Spec) ContainersInput {
	return newContainersInputForSources(containers, true)
}
