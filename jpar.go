// jpar is the Joint Photographic Archiver
package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"

	"github.com/bgilmore/jpar/archive"
)

func main() {
	if os.Args[1] == "dump" {
		fmt.Fprintf(os.Stderr, "dump returned %v\n", dump(os.Args[2]))
		return
	}

	w, err := archive.NewWriter(os.Stdout, 9)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Oh no:", err)
		return
	}

	cwd := os.DirFS(".")
	for _, name := range os.Args[1:] {
		f, err := cwd.Open(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, "open:", err)
			return
		}

		fi, err := f.Stat()
		if err != nil {
			fmt.Fprintln(os.Stderr, "stat:", err)
			return
		}

		header, err := tar.FileInfoHeader(fi, name)
		if err != nil {
			fmt.Fprintln(os.Stderr, "header:", err)
			return
		}

		if err := w.WriteHeader(header); err != nil {
			fmt.Fprintln(os.Stderr, "write header:", err)
			return
		}

		if _, err := io.Copy(w, f); err != nil {
			fmt.Fprintln(os.Stderr, "copy file:", err)
			return
		}

		f.Close()
	}

	if err := w.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "close??", err)
		return
	}
}

func dump(af string) error {
	f, err := os.Open(af)
	if err != nil {
		return err
	}

	r, err := archive.NewReader(f)
	if err != nil {
		return err
	}
	for {
		next, err := r.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		fi := next.FileInfo()
		fmt.Fprintf(os.Stdout, "%v\t0\t%s\t%s\t%d\t%v\t%v\n", fi.Mode(), next.Uname, next.Gname, next.Size, next.ModTime, next.Name)
	}
}

