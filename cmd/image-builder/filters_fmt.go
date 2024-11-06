package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/osbuild/images/pkg/imagefilter"
)

type FilteredResultFormatter interface {
	Output(io.Writer, []imagefilter.Result) error
}

func NewFilteredResultFormatter(format string) (FilteredResultFormatter, error) {
	switch format {
	case "", "text":
		return &textFilteredResultFormatter{}, nil
	case "json":
		return &jsonFilteredResultFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported formatter %q", format)
	}
}

type textFilteredResultFormatter struct{}

func (*textFilteredResultFormatter) Output(w io.Writer, all []imagefilter.Result) error {
	var errs []error
	for _, res := range all {
		if _, err := fmt.Fprintf(w, "%s arch:%s type:%s\n", res.Distro.Name(), res.Arch.Name(), res.ImgType.Name()); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

type jsonFilteredResultFormatter struct{}

// XXX: consider adding more details to the output
type filteredResultJSON struct {
	Distro struct {
		Name string `json:"name"`
	} `json:"distro"`
	Arch struct {
		Name string `json:"name"`
	} `json:"arch"`
	ImgType struct {
		Name string `json:"name"`
	} `json:"image_type"`
}

func (*jsonFilteredResultFormatter) Output(w io.Writer, all []imagefilter.Result) error {
	var out []filteredResultJSON

	for _, res := range all {
		out = append(out, filteredResultJSON{
			Distro: struct {
				Name string `json:"name"`
			}{
				Name: res.Distro.Name(),
			},
			Arch: struct {
				Name string `json:"name"`
			}{
				Name: res.Arch.Name(),
			},
			ImgType: struct {
				Name string `json:"name"`
			}{
				Name: res.ImgType.Name(),
			},
		})
	}

	enc := json.NewEncoder(w)
	return enc.Encode(out)
}
