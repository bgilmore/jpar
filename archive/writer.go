package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"time"
)

const framingVersion = 0x01

type Writer struct {
	*tar.Writer
	w  io.Writer
	sw io.WriteCloser
	zw *gzip.Writer
}

func NewWriter(w io.Writer, level int) (*Writer, error) {
	writer := &Writer{w: w}

	// Emit the JPEG prefix before piling on further framing.
	if _, err := writer.w.Write([]byte{0xff, 0xd8}); err != nil {
		return nil, err
	}

	// Wrap the base writer in a JPEG segment framer.
	writer.sw = NewSegmentWriter(writer.w)

	// Wrap the JPEG segment framer in a Gzip compression layer.
	// We abuse the SegmentWriter's Close() behavior to emit the Gzip header in
	// a standalone segment for versioning, etc.
	var err error
	writer.zw, err = gzip.NewWriterLevel(writer.sw, level)
	if err != nil {
		return nil, err
	}
	writer.zw.Header.ModTime = time.Now()
	writer.zw.Header.Extra = []byte{framingVersion}
	if err := writer.zw.Flush(); err != nil {
		return nil, err
	}
	writer.sw.Close() // HACK: this *should* be "Flush"

	// Finally wrap the Gzip writer with the tar writer.
	writer.Writer = tar.NewWriter(writer.zw)
	return writer, nil
}

func (w *Writer) Close() error {
	if err := w.Writer.Close(); err != nil {
		return err
	}
	if err := w.zw.Close(); err != nil {
		return err
	}
	return w.sw.Close()
}
