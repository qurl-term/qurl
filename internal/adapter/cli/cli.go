// Package cli is qURL's command-line adapter.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/qurl-term/qurl/internal/platform/buildinfo"
	"github.com/spf13/cobra"
)

type IO struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

const (
	exitOK    = 0
	exitErr   = 1
	exitUsage = 64
)

type usageError struct{ err error }

func (e usageError) Error() string { return e.err.Error() }
func (e usageError) Unwrap() error { return e.err }

func Execute(ctx context.Context, args []string, streams IO) int {
	root := newRootCmd(streams)
	// Non-nil so cobra never falls back to os.Args
	root.SetArgs(append([]string{}, args...))
	err := root.ExecuteContext(ctx)
	if err == nil {
		return exitOK
	}
	_, _ = fmt.Fprintf(streams.Err, "qurl: %v\n", err)
	var ue usageError
	if errors.As(err, &ue) {
		_, _ = fmt.Fprintln(streams.Err, "Run 'qurl --help' for usage.")
		return exitUsage
	}
	return exitErr
}

func newRootCmd(streams IO) *cobra.Command {
	info := buildinfo.Get()
	root := &cobra.Command{
		Use:           "qurl",
		Short:         "A fast terminal HTTP client",
		Version:       info.String(),
		Args:          noArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Replaced by the TUI later
			return cmd.Help()
		},
	}

	root.SetIn(streams.In)
	root.SetOut(streams.Out)
	root.SetErr(streams.Err)
	root.SetVersionTemplate("{{.Version}}\n")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageError{err: err} })
	root.CompletionOptions.DisableDefaultCmd = true
	root.Flags().Bool("version", false, "print version information and exit")
	root.AddCommand(newVersionCmd(info))
	return root
}

func newVersionCmd(info buildinfo.Info) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), info.String())
			return err
		},
	}
}

func noArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.NoArgs(cmd, args); err != nil {
		return usageError{err: err}
	}
	return nil
}
