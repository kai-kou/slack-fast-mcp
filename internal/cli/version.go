package cli

import (
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
		fmt.Fprintf(cmd.OutOrStdout(), `{"version":"%s","commit":"%s","date":"%s","go_version":"%s","platform":"%s/%s"}`+"\n",
			Version, Commit, Date, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "slack-fast-mcp %s\n", Version)
	fmt.Fprintf(cmd.OutOrStdout(), "  Commit:   %s\n", Commit)
	fmt.Fprintf(cmd.OutOrStdout(), "  Date:     %s\n", Date)
	fmt.Fprintf(cmd.OutOrStdout(), "  Go:       %s\n", runtime.Version())
	fmt.Fprintf(cmd.OutOrStdout(), "  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return nil
}
