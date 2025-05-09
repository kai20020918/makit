// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	mode     string
	timestamp string
	noCreate bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "makit [OPTION] <FILES|DIRS...>",
	Short: "Create files and directories with optional mode and timestamp",
	Long: `makit is a CLI tool to create directories and files 
with optional permissions, timestamps, and parent creation.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 今はオプションとターゲットを表示するだけ
		fmt.Println("Targets:", args)
		fmt.Println("Mode:", mode)
		fmt.Println("Timestamp:", timestamp)
		fmt.Println("NoCreate:", noCreate)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "", "Set file/directory mode (e.g. 755)")
	rootCmd.Flags().StringVarP(&timestamp, "date", "d", "", "Set timestamp (e.g. 202504181200)")
	rootCmd.Flags().BoolVarP(&noCreate, "no-create", "c", false, "Do not create if not exists")
}
