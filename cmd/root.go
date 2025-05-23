// cmd/root.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var (
	mode      string
	timestamp string
	noCreate  bool
)

var rootCmd = &cobra.Command{
	Use:   "makit [OPTION] <FILES|DIRS...>",
	Short: "Create files and directories with optional mode and timestamp",
	Long:  `makit is a CLI tool to create directories and files with optional permissions, timestamps, and parent creation.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var perm os.FileMode = 0755
		if mode != "" {
			parsed, err := strconv.ParseUint(mode, 8, 32)
			if err == nil {
				perm = os.FileMode(parsed)
			} else {
				fmt.Fprintf(os.Stderr, "invalid mode: %v\n", err)
				os.Exit(1)
			}
		}

		var tsTime time.Time
		if timestamp != "" {
			t, err := time.Parse("200601021504", timestamp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid timestamp format: %v\n", err)
				os.Exit(1)
			}
			tsTime = t
		}

		for _, path := range args {
			_, err := os.Stat(path)
			if os.IsNotExist(err) {
				if noCreate {
					fmt.Printf("Skipped (not created): %s\n", path)
					continue
				}

				if filepath.Ext(path) == "" {
					err := os.MkdirAll(path, perm)
					if err != nil {
						fmt.Printf("Error creating directory: %v\n", err)
						continue
					}
					fmt.Printf("Created directory: %s\n", path)
				} else {
					os.MkdirAll(filepath.Dir(path), perm)
					f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, perm)
					if err != nil {
						fmt.Printf("Error creating file: %s\n", err)
						continue
					}
					f.Close()
					fmt.Printf("Created file: %s\n", path)
				}
			} else {
				fmt.Printf("Exists: %s\n", path)
			}

			if mode != "" {
				os.Chmod(path, perm)
			}
			if !tsTime.IsZero() {
				os.Chtimes(path, tsTime, tsTime)
			}
		}
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
