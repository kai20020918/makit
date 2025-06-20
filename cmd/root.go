package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	// pflag をインポートに追加 (もしあれば)
)

var (
	mode      string
	timestamp string
	noCreate  bool
	verbose   bool
)

// newRootCommand は新しい Cobra コマンドインスタンスを作成し、フラグを初期化します。
// これにより、各テストが独立したコマンドインスタンスを持つことができます。
func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "makit [OPTION] <FILES|DIRS...>",
		Short: "Create files and directories with optional mode and timestamp",
		Long:  `makit is a CLI tool to create directories and files with optional permissions, timestamps, and parent creation.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// ★既存のRunロジックはここに含まれます
			if verbose {
				fmt.Println("Starting makit operation in verbose mode.")
				fmt.Printf("Arguments: %v\n", args)
				fmt.Printf("Mode: %s, Timestamp: %s, NoCreate: %t\n", mode, timestamp, noCreate)
			}

			var perm os.FileMode = 0755
			if mode != "" {
				parsed, err := strconv.ParseUint(mode, 8, 32)
				if err == nil {
					perm = os.FileMode(parsed)
				} else {
					fmt.Fprintf(os.Stderr, "invalid mode: %v\n", err)
					os.Exit(1) // エラー時は終了
				}
				if verbose {
					fmt.Printf("Parsed permission mode: %o\n", perm)
				}
			}

			var tsTime time.Time
			if timestamp != "" {
				t, err := time.Parse("200601021504", timestamp)
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid timestamp format: %v\n", err)
					os.Exit(1) // エラー時は終了
				}
				tsTime = t
				if verbose {
					fmt.Printf("Parsed timestamp: %s\n", tsTime.Format("2006-01-02 15:04:05"))
				}
			}

			for _, path := range args {
				if verbose {
					fmt.Printf("Processing path: %s\n", path)
				}

				_, err := os.Stat(path)
				if os.IsNotExist(err) {
					if noCreate {
						fmt.Printf("Skipped (not created): %s\n", path)
						if verbose {
							fmt.Println("Path does not exist and --no-create is set.")
						}
						continue
					}

					if filepath.Ext(path) == "" { // ディレクトリと判断
						if verbose {
							fmt.Printf("Path '%s' is identified as a directory. Creating...\n", path)
						}
						err := os.MkdirAll(path, perm)
						if err != nil {
							fmt.Printf("Error creating directory: %v\n", err)
							os.Exit(1) // エラー時は終了
						}
						fmt.Printf("Created directory: %s\n", path)
					} else { // ファイルと判断
						if verbose {
							fmt.Printf("Path '%s' is identified as a file. Creating...\n", path)
						}
						dir := filepath.Dir(path)
						if dir != "." {
							if verbose {
								fmt.Printf("Ensuring parent directory exists: %s\n", dir)
							}
							os.MkdirAll(dir, perm) // 親ディレクトリもデフォルトパーミッションで作成
						}
						f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, perm)
						if err != nil {
							fmt.Printf("Error creating file: %v\n", err)
							os.Exit(1) // エラー時は終了
						}
						f.Close()
						fmt.Printf("Created file: %s\n", path)
					}
				} else { // パスが存在する場合
					fmt.Printf("Exists: %s\n", path)
					if verbose {
						fmt.Println("Path already exists. Applying mode/timestamp if specified.")
					}
				}

				if mode != "" {
					if verbose {
						fmt.Printf("Applying mode %o to %s\n", perm, path)
					}
					os.Chmod(path, perm)
				}
				if !tsTime.IsZero() {
					if verbose {
						fmt.Printf("Applying timestamp %s to %s\n", tsTime.Format("2006-01-02 15:04:05"), path)
					}
					os.Chtimes(path, tsTime, tsTime)
				}
			}
		},
	}

    // ★以下を追加 (テスト時に不要な出力抑制のため)
    cmd.SilenceUsage = true
    cmd.SilenceErrors = true

	// ★新しいコマンドインスタンスに対してフラグを登録します
	cmd.Flags().StringVarP(&mode, "mode", "m", "", "Set file/directory mode (e.g. 755)")
	cmd.Flags().StringVarP(&timestamp, "date", "d", "", "Set timestamp (e.g. 202504181200)")
	cmd.Flags().BoolVarP(&noCreate, "no-create", "c", false, "Do not create if not exists")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}

func Execute() {
	// 実際のアプリケーション実行時には、ここで新しいコマンドインスタンスを作成して実行します。
	cmd := newRootCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// init() 関数はもう不要です（フラグ登録は newRootCommand() に移動したため）
// func init() { ... }