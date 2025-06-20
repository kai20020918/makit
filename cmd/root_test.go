// cmd/root_test.go (修正版)
package cmd

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// MockExitCalledWith は、exitFunc が呼び出された場合のコードを保持します
var MockExitCalledWith int

// setupTestEnv を修正し、テスト中に exitFunc をモックする
func setupTestEnv(t *testing.T) (cleanup func(), cmdInstance *cobra.Command) {
	// テスト用の作業ディレクトリを作成
	testDir, err := ioutil.TempDir("", "makit-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// 現在の作業ディレクトリを保存し、テストディレクトリに移動
	originalDir, _ := os.Getwd()
	os.Chdir(testDir)

	// os.Exit をテスト用にモックする
	originalExit := exitFunc
	MockExitCalledWith = -1 // リセット
	exitFunc = func(code int) {
		MockExitCalledWith = code
		// テストの実行を停止するためにパニックを発生させる
		// 実際の os.Exit は呼び出さない
		panic("os.Exit called during test")
	}

	// 各テストのために新しいコマンドインスタンスを作成
	testCmd := newRootCommand()

	// クリーンアップ関数を返す
	cleanupFunc := func() {
		os.Chdir(originalDir) // 元のディレクトリに戻る
		os.RemoveAll(testDir) // テストディレクトリを削除
		exitFunc = originalExit // テスト終了後に元の os.Exit に戻す
	}

	return cleanupFunc, testCmd
}

// executeAndCaptureOutput はコマンドを実行し、標準出力と標準エラー出力をキャプチャします。
// Cobraコマンドインスタンスを引数として受け取るように変更
func executeAndCaptureOutput(t *testing.T, cmd *cobra.Command, args []string) (output string, err error) {
	// 標準出力と標準エラーをキャプチャするためのパイプを作成
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut // os.Stdout をパイプの書き込み側にリダイレクト
	os.Stderr = wErr // os.Stderr をパイプの書き込み側にリダイレクト

	// Cobraコマンドの引数を設定
	cmd.SetArgs(args)

	// defer-recover ブロックを使って、モックされた os.Exit からのパニックをキャッチする
	defer func() {
		if r := recover(); r != nil {
			if str, ok := r.(string); ok && strings.Contains(str, "os.Exit called during test") {
				// これは期待されるパニックなので、何もしない
			} else {
				panic(r) // 予期せぬパニックは再パニックさせる
			}
		}
		// パイプを閉じて、出力を読み込む
		wOut.Close()
		wErr.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		outBytes, _ := ioutil.ReadAll(rOut)
		errBytes, _ := ioutil.ReadAll(rErr)
		rOut.Close()
		rErr.Close()
		output = string(outBytes) + string(errBytes)
	}()

	err = cmd.Execute() // コマンドを実行する

	return output, err
}

// TestExecuteCreateFile はファイルの作成をテストします。
func TestExecuteCreateFile(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	testFileName := "test_file.txt"
	output, err := executeAndCaptureOutput(t, cmd, []string{testFileName})
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
	if !strings.Contains(output, expectedFileCreationOutput) {
		t.Errorf("Expected output %q not found in %q", expectedFileCreationOutput, output)
	}
}


// TestExecuteNoCreateExistingFile は -c オプションと既存ファイルをテストします。
func TestExecuteNoCreateExistingFile(t *testing.T) {
	cleanup, cmd := setupTestEnv(t)
	defer cleanup()

	testFileName := "existing.txt"
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
	
	// Execute() がエラーを返すか、または MockExitCalledWith が設定されているかを確認
	if err != nil {
        // Execute() がエラーを返した場合、ここでエラーメッセージをチェック
    } else if MockExitCalledWith != 1 {
		t.Errorf("Expected an error or os.Exit(1), but got none. MockExitCalledWith: %d", MockExitCalledWith)
	}

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

	if err != nil {
        // Execute() がエラーを返した場合、ここでエラーメッセージをチェック
    } else if MockExitCalledWith != 1 {
		t.Errorf("Expected an error or os.Exit(1), but got none. MockExitCalledWith: %d", MockExitCalledWith)
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
	loc, _ := time.LoadLocation("UTC") // タイムゾーンをUTCに明示的に設定
	expectedTime, _ := time.ParseInLocation("200601021504", timestampStr, loc)


	_, err := executeAndCaptureOutput(t, cmd, []string{"-d", timestampStr, testFileName})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	info, err := os.Stat(testFileName)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", testFileName, err)
	}
	// 最終更新時刻をUTCに変換し、秒単位で切り捨てて比較する
	if info.ModTime().In(loc).Truncate(time.Second) != expectedTime.Truncate(time.Second) {
		t.Errorf("Expected modification time %s (UTC), got %s (UTC)", expectedTime.Format(time.RFC3339), info.ModTime().In(loc).Format(time.RFC3339))
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
		if MockExitCalledWith != 1 {
			t.Errorf("Expected an error or os.Exit(1), but got none. MockExitCalledWith: %d", MockExitCalledWith)
		}
	}
	
	expectedErrorOutput1 := "Error: accepts at least 1 arg(s), received 0"
	expectedErrorOutput2 := "Error: requires at least 1 arg(s), only received 0"
	
	if !strings.Contains(output, expectedErrorOutput1) && !strings.Contains(output, expectedErrorOutput2) {
		t.Errorf("Expected error message %q or %q not found in %q", expectedErrorOutput1, expectedErrorOutput2, output)
	}
}