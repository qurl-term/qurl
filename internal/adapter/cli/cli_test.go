package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{name: "version", args: []string{"version"}, wantCode: exitOK, wantOut: "qurl "},
		{name: "version flag", args: []string{"--version"}, wantCode: exitOK, wantOut: "qurl "},
		{name: "no args prints help", wantCode: exitOK, wantOut: "Usage:"},
		{name: "unknown command", args: []string{"nope"}, wantCode: exitUsage, wantErr: `unknown command "nope"`},
		{name: "unknown flag", args: []string{"--nope"}, wantCode: exitUsage, wantErr: "unknown flag: --nope"},
		{name: "version rejects args", args: []string{"version", "extra"}, wantCode: exitUsage, wantErr: `unknown command "extra"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := Execute(t.Context(), tt.args, IO{In: strings.NewReader(""), Out: &out, Err: &errOut})
			if code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr: %q)", code, tt.wantCode, errOut.String())
			}
			if !strings.Contains(out.String(), tt.wantOut) {
				t.Errorf("stdout = %q, want it to contain %q", out.String(), tt.wantOut)
			}
			if !strings.Contains(errOut.String(), tt.wantErr) {
				t.Errorf("stderr = %q, want it to contain %q", errOut.String(), tt.wantErr)
			}
		})
	}
}
