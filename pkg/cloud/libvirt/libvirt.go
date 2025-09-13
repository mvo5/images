package libvirt

import (
	"fmt"
	"github.com/osbuild/images/pkg/cloud"
	"io"
	"os"

	lv "libvirt.org/go/libvirt"
)

type libvirtUploader struct {
	connection string
	pool       string
	volume     string
}

func NewUploader(connection string, pool string, volume string) (cloud.Uploader, error) {
	return &libvirtUploader{
		connection: connection,
		pool:       pool,
		volume:     volume,
	}, nil
}

func (lu *libvirtUploader) Check(status io.Writer) error {
	return nil
}

func (lu *libvirtUploader) UploadAndRegister(r io.Reader, f *os.File, status io.Writer) (err error) {
	fmt.Fprintf(status, "Uploading to libvirt...\n")

	// Connect to libvirt
	conn, err := lv.NewConnect(lu.connection)
	if err != nil {
		return fmt.Errorf("Failed to connect to libvirt: %v", err)
	}
	defer conn.Close()

	// Lookup storage pool
	pool, err := conn.LookupStoragePoolByName("default")
	if err != nil {
		return fmt.Errorf("Failed to find storage pool: %v", err)
	}
	defer pool.Free()

	// Get file info (for volume size)
	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("Failed to stat qcow2 file: %v", err)
	}

	// Create the volume in the pool
	volXML := lu.VolumeXML(lu.volume, stat.Size())
	vol, err := pool.StorageVolCreateXML(volXML, 0)
	if err != nil {
		return fmt.Errorf("Failed to create a libvirt volume: %v", err)
	}
	defer vol.Free()

	// Upload data into the volume
	err = lu.Upload(conn, vol, r, stat.Size())
	if err != nil {
		return fmt.Errorf("Failed to upload the file to libvirt: %v", err)
	}

	fmt.Fprintf(status, "File %s uploaded to %s\n", f.Name(), lu.connection)
	return nil
}

func (lu *libvirtUploader) VolumeXML(name string, size int64) string {
	return fmt.Sprintf(`
<volume>
  <name>%s</name>
  <capacity unit="bytes">%d</capacity>
  <target>
	<format type="qcow2"/>
  </target>
</volume>`, name, size)
}

func (lu *libvirtUploader) Upload(conn *lv.Connect, vol *lv.StorageVol, r io.Reader, size int64) (err error) {
	// Initialize the upload stream
	stream, err := conn.NewStream(lv.STREAM_NONBLOCK)
	if err != nil {
		return fmt.Errorf("Failed to initialize an upload stream: %v", err)
	}
	defer stream.Free()

	// Start the upload
	if err := vol.Upload(stream, 0, uint64(size), 0); err != nil {
		return fmt.Errorf("Failed to start the upload: %v", err)
	}

	// Stream the content using a 64KB buffer
	buf := make([]byte, 64*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, sendErr := stream.Send(buf[:n]); sendErr != nil {
				return fmt.Errorf("Failed to stream the buffer: %v", sendErr)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("Failed to read the file: %v", err)
		}
	}

	// Finish the stream
	if err := stream.Finish(); err != nil {
		return fmt.Errorf("Failed to finish stream: %v", err)
	}

	return nil
}
