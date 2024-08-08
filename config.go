package main

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/multierr"
)

type Mode uint8

const (
	ModeNewfile = Mode(1)
)

var RootCmd cobra.Command

func init() {
	RootCmd = cobra.Command{
		Use:  "dos2unix - DOS/MAC to UNIX text file format converter",
		RunE: RunE,
	}

	flags := RootCmd.Flags()
	flags.BoolP("quiet", "q", false, "Quiet Mode. Suppress all warning and messages") // sic
	flags.BoolP("keepdate", "k", false, "Keep the date stamp of output file same as input file.")
	flags.BoolP("convmode", "c", false, "Kept for backwards compatibility with dos2unix")
	flags.BoolP("oldfile", "o", false, "Old file mode. Convert the file and write output to it. The program default to run in this mode. Wildcard names may be used.")
	flags.BoolP("newfile", "n", false, "New file mode. Convert the infile and write output to outfile. File names must be given in pairs and wildcard names should NOT be used or you WILL lose your files.")
}

type config struct {
	quiet    bool
	keepdate bool
	mode     Mode
	files    []string
}

func ParseFlags(flags *pflag.FlagSet) (*config, error) {
	cfg := config{}
	var merr error
	var err error

	if cfg.quiet, err = flags.GetBool("quiet"); err != nil {
		merr = multierr.Append(merr, err)
	}

	if cfg.keepdate, err = flags.GetBool("keepdate"); err != nil {
		merr = multierr.Append(merr, err)
	}

	if isNewfile, err := flags.GetBool("newfile"); err != nil {
		merr = multierr.Append(merr, err)
	} else if isNewfile {
		cfg.mode = ModeNewfile
	}

	if cfg.mode == ModeNewfile {
		if val, err := flags.GetBool("oldfile"); err != nil {
			merr = multierr.Append(merr, err)
		} else if val {
			merr = multierr.Append(merr, errors.New("cannot specify oldfile and newfile simultaneously"))
		}
	}

	if cfg.mode == ModeNewfile {
		cfg.files, err = expandFilesNewfile(flags.Args())
		merr = multierr.Append(merr, err)
	} else {
		cfg.files, err = expandFilesDefault(flags.Args())
		merr = multierr.Append(merr, err)
	}

	return &cfg, merr
}

func expandFilesDefault(files []string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}

	retval := make([]string, 0, len(files))
	var merr error

	for _, fn := range files {
		m, err := filepath.Glob(fn)
		if err != nil {
			merr = multierr.Append(merr, err)
			continue
		}
		retval = append(retval, m...)
	}

	return retval, merr
}

func expandFilesNewfile(files []string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}

	retval := make([]string, 0, len(files))
	var merr error

	for _, fn := range files {
		if strings.Contains(fn, "*") {
			merr = multierr.Append(merr, errors.New("Cannot use globs in newfile mode: "+fn))
		}

		retval = append(retval, fn)
	}

	if 1 == len(retval)%2 {
		merr = multierr.Append(merr, errors.New("newfile mode requires pairs of filenames"))
	}

	if 0 == len(retval) {
		merr = multierr.Append(merr, errors.New("newfile mode requires at least one pair of filenames"))
	}

	return retval, merr
}
