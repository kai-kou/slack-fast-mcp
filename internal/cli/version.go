package cli

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// newVersionCmd は version サブコマンドを作成する。
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display the version, Go version, and platform information.",
		RunE:  runVersion,
	}
}

// runVersion はバージョン情報を出力する。
func runVersion(cmd *cobra.Command, args []string) error {
	if flagJSON {
		out := map[string]string{
			"version":    Version,
			"commit":     Commit,
			"date":       Date,
			"go_version": runtime.Version(),
			"platform":   runtime.GOOS + "/" + runtime.GOARCH,
		}
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetEscapeHTML(false)
		return encoder.Encode(out)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "slack-fast-mcp %s\n", Version)
	fmt.Fprintf(cmd.OutOrStdout(), "  Commit:   %s\n", Commit)
	fmt.Fprintf(cmd.OutOrStdout(), "  Date:     %s\n", Date)
	fmt.Fprintf(cmd.OutOrStdout(), "  Go:       %s\n", runtime.Version())
	fmt.Fprintf(cmd.OutOrStdout(), "  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return nil
}
