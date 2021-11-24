package archive

import (
	"bytes"
	"errors"
	"io"
)

const (
	segmentPrefixLength = 7
	maxSegmentLength    = 0xffff - segmentPrefixLength
)

var (
	ErrUnrecognizedSegment = errors.New("archive: reached unrecognized segment")
)

type segmentWriter struct {
	w   io.Writer
	buf bytes.Buffer
}

func NewSegmentWriter(w io.Writer) io.WriteCloser {
	return &segmentWriter{w: w}
}

func (w *segmentWriter) flushSegment() error {
	len := w.buf.Len() + segmentPrefixLength
	if _, err := w.w.Write([]byte{0xff, 0xe0, byte(len >> 8), byte(len), 'J', 'P', 'A', 'R', 0}); err != nil {
		return err
	}
	if _, err := w.buf.WriteTo(w.w); err != nil {
		return err
	}
	w.buf.Reset()
	return nil
}

func (w *segmentWriter) Write(p []byte) (n int, err error) {
	ofs := 0
	for {
		chunk := p[ofs:]
		space := maxSegmentLength - w.buf.Len()
		if len(chunk) > space {
			chunk = chunk[:space]
		}

		n, err := w.buf.Write(chunk)
		ofs += n
		if err != nil {
			return ofs, err
		}

		if w.buf.Len() == maxSegmentLength {
			if err := w.flushSegment(); err != nil {
				return ofs, err
			}
		}

		if len(p)-ofs == 0 {
			return ofs, nil
		}
	}
}

func (w *segmentWriter) Close() error {
	if w.buf.Len() > 0 {
		w.flushSegment()
	}
	return nil
}

type segmentReader struct {
	r   io.Reader
	buf bytes.Buffer
}

func NewSegmentReader(r io.Reader) io.Reader {
	return &segmentReader{r: r}
}

func (r *segmentReader) bufferSegment() error {
	// Check whether the next segment is an APP0 segment.
	marker := make([]byte, 2)
	if _, err := io.ReadFull(r.r, marker); err != nil {
		return err
	}
	if marker[0] != 0xff || marker[1] != 0xe0 {
		return ErrUnrecognizedSegment
	}

	// Verify that the APP0 segment was written by JPAR.
	header := make([]byte, segmentPrefixLength)
	if _, err := io.ReadFull(r.r, header); err != nil {
		return err
	}
	if bytes.Compare(header[2:], []byte{'J', 'P', 'A', 'R', 0}) != 0 {
		return ErrUnrecognizedSegment
	}

	// Buffer the segment data.
	length := (uint16(header[1]) | uint16(header[0])<<8) - segmentPrefixLength
	r.buf.Reset()
	if _, err := io.CopyN(&r.buf, r.r, int64(length)); err != nil {
		return err
	}

	return nil
}

func (r *segmentReader) Read(p []byte) (n int, err error) {
	if r.buf.Len() == 0 {
		if err := r.bufferSegment(); err != nil {
			return 0, err
		}
	}
	return r.buf.Read(p)
}
