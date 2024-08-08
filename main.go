package main

import (
	"bufio"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"go.uber.org/multierr"
)

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func RunE(cmd *cobra.Command, args []string) error {
	cfg, err := ParseFlags(RootCmd.Flags())
	if err != nil {
		return err
	}

	if len(cfg.files) == 0 {
		return RunStdin()
	}

	if cfg.mode == ModeNewfile {
		return RunFilesNewfile(cfg.files, cfg.keepdate)
	}

	return RunFilesDefault(cfg.files, cfg.keepdate)
}

func RunStdin() error {
	w := bufio.NewWriterSize(os.Stdout, defaultBufferSize)
	r := bufio.NewReaderSize(os.Stdin, defaultBufferSize)
	if _, err := dos2unix64(w, r); err != nil {
		return err
	}
	return w.Flush()
}

func RunFilesDefault(files []string, keepdate bool) error {
	w := bufio.NewWriterSize(nil, defaultBufferSize)
	r := bufio.NewReaderSize(nil, defaultBufferSize)

	pid := os.Getpid()

	for i, fname := range files {
		fi, err := os.Stat(files[0])
		if err != nil {
			return err
		}
		ts := fi.ModTime()

		tmpFname := path.Join(os.TempDir(), fmt.Sprintf("dos2unix.%06d.%06d", pid, i))
		outfile, err := os.Create(tmpFname)
		if err != nil {
			return err
		}

		infile, err := os.Open(fname)
		if err != nil {
			_ = outfile.Close()
			return err
		}

		w.Reset(outfile)
		r.Reset(infile)
		_, err = dos2unix64(w, r)
		err = multierr.Combine(
			err,
			infile.Close(),
			w.Flush(),
			outfile.Close(),
		)
		if err != nil {
			return err
		}

		if err := os.Rename(tmpFname, fname); err != nil {
			return err
		}
		if keepdate {
			if err := os.Chtimes(fname, ts, ts); err != nil {
				return err
			}
		}
	}
	return nil
}

func RunFilesNewfile(files []string, keepdate bool) error {
	w := bufio.NewWriterSize(nil, defaultBufferSize)
	r := bufio.NewReaderSize(nil, defaultBufferSize)

	for i := 0; i < len(files); i += 2 {
		fi, err := os.Stat(files[0])
		if err != nil {
			return err
		}
		ts := fi.ModTime()

		outfile, err := os.Create(files[1])
		if err != nil {
			return err
		}

		infile, err := os.Open(files[0])
		if err != nil {
			_ = outfile.Close()
			return err
		}

		w.Reset(outfile)
		r.Reset(infile)
		_, err = dos2unix64(w, r)
		if err := multierr.Combine(
			err,
			infile.Close(),
			w.Flush(),
			outfile.Close(),
		); err != nil {
			return err
		}
		if keepdate {
			if err := os.Chtimes(files[1], ts, ts); err != nil {
				return err
			}
		}
	}
	return nil
}
