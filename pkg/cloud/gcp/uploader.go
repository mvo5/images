package gcp

import (
	"bytes"
	"context"
	// gcp uses MD5 hashes
	/* #nosec G501 */
	"crypto/md5"
	"errors"
	"fmt"
	"io"

	"cloud.google.com/go/compute/apiv1/computepb"

	"github.com/osbuild/images/pkg/cloud"
)

var _ = cloud.Uploader(&gcpUploader{})

type gcpUploader struct {
	client *GCP

	region          string
	bucketName      string
	imageName       string
	guestOSFeatures []*computepb.GuestOsFeature
}

func (g *gcpUploader) Check(io.Writer) error {
	// XXX: add check for bucket write permission
	return nil
}

func (g *gcpUploader) UploadAndRegister(r io.Reader, status io.Writer) (err error) {
	ctx := context.TODO()

	imageFileHash := md5.New()
	tr := io.TeeReader(r, imageFileHash)

	fmt.Fprintf(status, "Uploading %s to %s\n", g.imageName, g.bucketName)
	imgAttrs, err := g.client.StorageObjectUploadFromReader(ctx, tr, nil, g.bucketName, g.imageName, map[string]string{MetadataKeyImageName: g.imageName})
	if err != nil {
		return err
	}
	if !bytes.Equal(imgAttrs.MD5, imageFileHash.Sum(nil)) {
		return fmt.Errorf("upload failed: md5sum mismatch %q != %q", imgAttrs.MD5, imageFileHash.Sum(nil))
	}

	fmt.Fprintf(status, "File uploaded\n")
	defer func() {
		if err != nil {
			fmt.Fprintf(status, "Cleaning up GCP object %s:%s\n", g.bucketName, g.imageName)

			delErr := g.client.StorageObjectDelete(ctx, g.bucketName, g.imageName)
			err = errors.Join(err, delErr)
		}
	}()

	fmt.Fprintf(status, "Registering compute image %s\n", g.imageName)
	if _, err = g.client.ComputeImageInsert(ctx, g.bucketName, g.imageName, g.imageName, []string{g.region}, g.guestOSFeatures); err != nil {
		return err
	}
	fmt.Fprintf(status, "Deleting temporary storage object %s:%s\n", g.bucketName, g.imageName)
	if err := g.client.StorageObjectDelete(ctx, g.bucketName, g.imageName); err != nil {
		return err
	}

	fmt.Fprintf(status, "Computer image registered as %s\n", g.client.ComputeImageURL(g.imageName))

	return nil
}

// note that the distroName is the target OS, e.g. "centos-9"
func NewUploader(region, bucket, imageName, distroName string) (cloud.Uploader, error) {
	client, err := New(nil)
	if err != nil {
		return nil, err
	}

	return &gcpUploader{
		client:          client,
		region:          region,
		bucketName:      bucket,
		imageName:       imageName,
		guestOSFeatures: GuestOsFeaturesByDistro(distroName),
	}, nil
}
