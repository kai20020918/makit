package cmd

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// setupTestEnv はテスト環境をセットアップし、終了時にクリーンアップします。
// 各テストで新しいコマンドインスタンスを返します。
func setupTestEnv(t *testing.T) (cleanup func(), cmdInstance *cobra.Command) {
	// テスト用の作業ディレクトリを作成
	testDir, err := ioutil.TempDir("", "makit-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// 現在の作業ディレクトリを保存し、テストディレクトリに移動
	originalDir, _ := os.Getwd()
	os.Chdir(testDir)

	// 各テストのために新しいコマンドインスタンスを作成
	testCmd := newRootCommand() // ★ここが重要：新しいコマンドインスタンスを取得

	// クリーンアップ関数を返す
	cleanupFunc := func() {
		os.Chdir(originalDir) // 元のディレクトリに戻る
		os.RemoveAll(testDir) // テストディレクトリを削除

		// グローバルな os.Stdout/Stderr のリダイレクトをテストごとに元に戻す
		// Cobra の SetOut/SetErr はここでは行わない (Execute()内で処理されるため)
	}

	return cleanupFunc, testCmd
}

// executeAndCaptureOutput はコマンドを実行し、標準出力と標準エラー出力をキャプチャします。
// Cobraコマンドインスタンスを引数として受け取るように変更
func executeAndCaptureOutput(t *testing.T, cmd *cobra.Command, args []string) (string, error) {
	// 標準出力と標準エラーをキャプチャするためのパイプを作成
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut // os.Stdout をパイプの書き込み側にリダイレクト
	os.Stderr = wErr // os.Stderr をパイプの書き込み側にリダイレクト

	// Cobraコマンドの引数を設定
	cmd.SetArgs(args)

	// Execute() を呼び出し
	// 注意: Execute() が os.Exit() を呼ぶ可能性があるため、
	// テスト内で直接 os.Exit() を防ぐためには、
	// rootCmd.ExecuteC() を使うのがより頑健な方法です。
	// しかし、ここでは現状のコードに合わせて Execute() を使います。
	// cmd.Execute() がエラーを返さず os.Exit(1) を呼んだ場合、テストが失敗します。
	err := cmd.Execute()

	// パイプを閉じて、出力を読み込む準備
	wOut.Close()
	wErr.Close()
	
	// 標準出力と標準エラーを元に戻す (重要)
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// 出力を読み込む
	outBytes, _ := ioutil.ReadAll(rOut)
	errBytes, _ := ioutil.ReadAll(rErr)

	// バッファを閉じる
	rOut.Close()
	rErr.Close()

	// 標準出力と標準エラーの文字列を結合して返す
	return string(outBytes) + string(errBytes), err
}

// TestExecuteCreateFile はファイルの作成をテストします。
func TestExecuteCreateFile(t *testing.T) {
	cleanup, cmd := setupTestEnv(t) // ★ cmd インスタンスを取得
	defer cleanup()

	testFileName := "test_file.txt"
	output, err := executeAndCaptureOutput(t, cmd, []string{testFileName}) // ★ cmd を渡す
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if _, err := os.Stat(testFileName); os.IsNotExist(err) {
		t.Errorf("File %s was not created", testFileName)
	}
	expectedOutput := "Created file: " + testFileName + "\n"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

// TestExecuteCreateDirectory はディレクトリの作成をテストします。
func TestExecuteCreateDirectory(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	testDirName := "test_directory/"
	output, err := executeAndCaptureOutput(t, cmd, []string{testDirName})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if _, err := os.Stat(testDirName); os.IsNotExist(err) {
		t.Errorf("Directory %s was not created", testDirName)
	}
	expectedOutput := "Created directory: " + testDirName + "\n"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

// TestExecuteCreateNestedDirectory はネストされたディレクトリの作成をテストします。
func TestExecuteCreateNestedDirectory(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	testPath := "parent/child/grandchild/"
	output, err := executeAndCaptureOutput(t, cmd, []string{testPath})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("Nested directory %s was not created", testPath)
	}
	expectedOutput := "Created directory: " + testPath + "\n"
	// MkdirAllのverbose出力が追加されたので、strings.Containsがより適切になる可能性あり
	// ここは簡潔さを保つため Contains でよい
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output %q not found in %q", expectedOutput, output)
	}
}

// TestExecuteCreateFileInNewDirectory は新しいディレクトリ内にファイルを作成するテストです。
func TestExecuteCreateFileInNewDirectory(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	testPath := "new_dir/new_file.txt"
	output, err := executeAndCaptureOutput(t, cmd, []string{testPath})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("File %s was not created in new directory", testPath)
	}
	expectedFileCreationOutput := "Created file: " + testPath + "\n"
	// MkdirAll の出力があるかもしれないので Contains を使用
	if !strings.Contains(output, expectedFileCreationOutput) {
		t.Errorf("Expected output %q not found in %q", expectedFileCreationOutput, output)
	}
}


// TestExecuteNoCreateExistingFile は -c オプションと既存ファイルをテストします。
func TestExecuteNoCreateExistingFile(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	testFileName := "existing.txt"
	// 事前にファイルを作成
	err := ioutil.WriteFile(testFileName, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("Failed to create pre-existing file: %v", err)
	}

	output, cmdErr := executeAndCaptureOutput(t, cmd, []string{"-c", testFileName})
	if cmdErr != nil {
		t.Fatalf("Execute() failed: %v", cmdErr)
	}

	expectedOutput := "Exists: " + testFileName + "\n"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}

	// ファイルの内容が変更されていないことを確認 (内容変更機能はないので)
	content, _ := ioutil.ReadFile(testFileName)
	if string(content) != "hello" {
		t.Errorf("File content was unexpectedly modified")
	}
}

// TestExecuteNoCreateNonExistingFile は -c オプションと存在しないファイルをテストします。
func TestExecuteNoCreateNonExistingFile(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	testFileName := "non_existing.txt"
	output, cmdErr := executeAndCaptureOutput(t, cmd, []string{"-c", testFileName})
	if cmdErr != nil {
		t.Fatalf("Execute() failed: %v", cmdErr)
	}

	if _, err := os.Stat(testFileName); !os.IsNotExist(err) {
		t.Errorf("File %s was unexpectedly created", testFileName)
	}
	expectedOutput := "Skipped (not created): " + testFileName + "\n"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

// TestExecuteInvalidMode は無効なモード指定をテストします。
func TestExecuteInvalidMode(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	output, err := executeAndCaptureOutput(t, cmd, []string{"-m", "invalid_mode", "dummy.txt"})
	if err == nil {
		t.Errorf("Expected an error for invalid mode, but got none")
	}

	// エラーは stderr に出力される可能性もあるので、両方をキャプチャして確認
	expectedErrorOutput := "invalid mode"
	if !strings.Contains(output, expectedErrorOutput) {
		t.Errorf("Expected error message %q not found in %q", expectedErrorOutput, output)
	}
}

// TestExecuteValidMode は有効なモード指定をテストします。
func TestExecuteValidMode(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	testFileName := "mode_test.txt"
	modeStr := "755"
	expectedPerm := os.FileMode(0755)

	_, err := executeAndCaptureOutput(t, cmd, []string{"-m", modeStr, testFileName})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	info, err := os.Stat(testFileName)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", testFileName, err)
	}
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("Expected permission %o, got %o", expectedPerm, info.Mode().Perm())
	}
}

// TestExecuteInvalidTimestamp は無効なタイムスタンプ指定をテストします。
func TestExecuteInvalidTimestamp(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	output, err := executeAndCaptureOutput(t, cmd, []string{"-d", "invalid_timestamp", "dummy.txt"})
	if err == nil {
		t.Errorf("Expected an error for invalid timestamp, but got none")
	}

	expectedErrorOutput := "invalid timestamp format"
	if !strings.Contains(output, expectedErrorOutput) {
		t.Errorf("Expected error message %q not found in %q", expectedErrorOutput, output)
	}
}

// TestExecuteValidTimestamp は有効なタイムスタンプ指定をテストします。
func TestExecuteValidTimestamp(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	testFileName := "timestamp_test.txt"
	timestampStr := "202401011030"
	expectedTime, _ := time.Parse("200601021504", timestampStr)

	_, err := executeAndCaptureOutput(t, cmd, []string{"-d", timestampStr, testFileName})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	info, err := os.Stat(testFileName)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", testFileName, err)
	}
	// Chtimesはアクセス時間と変更時間を両方セットするので、変更時間を確認
	if info.ModTime().Truncate(time.Second) != expectedTime.Truncate(time.Second) {
		t.Errorf("Expected modification time %s, got %s", expectedTime, info.ModTime())
	}
}

// TestExecuteVerboseMode は詳細出力オプションをテストします。
func TestExecuteVerboseMode(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	testFileName := "verbose_file.txt"
	output, err := executeAndCaptureOutput(t, cmd, []string{"-v", testFileName})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if !strings.Contains(output, "Starting makit operation in verbose mode.") {
		t.Errorf("Verbose initial message not found")
	}
	if !strings.Contains(output, "Processing path: "+testFileName) {
		t.Errorf("Verbose processing message not found")
	}
	if !strings.Contains(output, "Created file: "+testFileName) {
		t.Errorf("File creation message not found")
	}
}

// TestExecuteMultipleArgs は複数の引数をテストします。
func TestExecuteMultipleArgs(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	args := []string{"file1.txt", "dir1/", "dir2/file2.txt"}
	output, err := executeAndCaptureOutput(t, cmd, args)
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	for _, arg := range args {
		if _, err := os.Stat(arg); os.IsNotExist(err) {
			t.Errorf("Path %s was not created", arg)
		}
		var expectedOutput string
		if strings.HasSuffix(arg, "/") {
			expectedOutput = "Created directory: " + arg + "\n"
		} else {
			expectedOutput = "Created file: " + arg + "\n"
		}
		if !strings.Contains(output, expectedOutput) {
			t.Errorf("Expected output %q for %s not found in %q", expectedOutput, arg, output)
		}
	}
}

// TestExecuteEmptyArgs は引数がない場合をテストします (CobraがMinimumNArgs(1)を設定しているためエラーになるはず)。
func TestExecuteEmptyArgs(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	output, err := executeAndCaptureOutput(t, cmd, []string{}) // 空の引数リストを渡す
	
	if err == nil {
		t.Errorf("Expected an error for no arguments, but got none")
	}
	// エラーメッセージの確認（Cobraのデフォルトエラーメッセージを期待）
	// Cobraのバージョンによってメッセージが異なる可能性があるのでContainsで確認
	expectedErrorOutput := "Error: accepts at least 1 arg(s), received 0"
	if !strings.Contains(output, expectedErrorOutput) {
		expectedErrorOutput = "Error: requires at least 1 arg(s), only received 0"
		if !strings.Contains(output, expectedErrorOutput) {
			t.Errorf("Expected error message %q or %q not found in %q", "Error: accepts at least 1 arg(s), received 0", "Error: requires at least 1 arg(s), only received 0", output)
		}
	}
}