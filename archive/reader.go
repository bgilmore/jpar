package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
)

type Reader struct {
	*tar.Reader
	r  io.Reader
	sr io.Reader
	zr *gzip.Reader
}

func NewReader(r io.Reader) (*Reader, error) {
	magic := make([]byte, 2)
	if _, err := r.Read(magic); err != nil {
		return nil, err
	}
	if magic[0] != 0xff || magic[1] != 0xd8 {
		return nil, fmt.Errorf("archive: missing JPEG SOI marker")
	}

	reader := &Reader{r: r}
	reader.sr = NewSegmentReader(reader.r)

	var err error
	reader.zr, err = gzip.NewReader(reader.sr)
	if err != nil {
		return nil, err
	}
	if len(reader.zr.Header.Extra) != 1 || reader.zr.Header.Extra[0] != framingVersion {
		return nil, fmt.Errorf("archive: unrecognized framing version %x", reader.zr.Header.Extra)
	}

	reader.Reader = tar.NewReader(reader.zr)
	return reader, nil
}
