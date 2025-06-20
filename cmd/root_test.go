// cmd/root_test.go
package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time" // time パッケージを追加
)

// setupTestEnv はテスト環境をセットアップし、終了時にクリーンアップします。
// テスト後に作成されたファイルを確実に削除するために使用します。
func setupTestEnv(t *testing.T) (cleanup func()) {
	// テスト用の作業ディレクトリを作成
	testDir, err := ioutil.TempDir("", "makit-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// 現在の作業ディレクトリを保存し、テストディレクトリに移動
	originalDir, _ := os.Getwd()
	os.Chdir(testDir)

	// 各テストの前にCobraコマンドの引数とフラグをリセット
	rootCmd.SetArgs(nil)
	rootCmd.ResetFlags() // フラグをリセット

	// Cobraの出力先もデフォルトに戻しておく (executeAndCaptureOutput でテストごとに設定し直されるため)
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	return func() {
		// テストディレクトリから元のディレクトリに戻る
		os.Chdir(originalDir)
		// テストディレクトリを削除
		os.RemoveAll(testDir)

		// ★Cobraの出力先を確実に元の状態に戻す（このcleanupが呼ばれる時点でテストは終了）
		rootCmd.SetOut(os.Stdout)
		rootCmd.SetErr(os.Stderr)
	}
}

// executeAndCaptureOutput はコマンドを実行し、標準出力と標準エラー出力をキャプチャします。
func executeAndCaptureOutput(t *testing.T, args []string) (string, error) {
	// 標準出力と標準エラーをキャプチャするためのパイプを作成
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut // os.Stdout をパイプの書き込み側にリダイレクト
	os.Stderr = wErr // os.Stderr をパイプの書き込み側にリダイレクト

	// Cobraコマンドの引数を設定
	rootCmd.SetArgs(args)

	// Execute() を呼び出し
	// 注意: Execute() は os.Exit() を呼ぶ可能性があるため、
	// テスト内で直接 os.Exit() を防ぐためには、
	// rootCmd.ExecuteC() を使うのがより頑健な方法ですが、
	// ここでは現状のコードに合わせて Execute() を使います。
	err := rootCmd.Execute()

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
	cleanup := setupTestEnv(t)
	defer cleanup()

	testFileName := "test_file.txt"
	output, err := executeAndCaptureOutput(t, []string{"makit", testFileName})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if _, err := os.Stat(testFileName); os.IsNotExist(err) {
		t.Errorf("File %s was not created", testFileName)
	}
	expectedOutput := "Created file: " + testFileName + "\n" // ★改行を追加
	if output != expectedOutput { // ★strings.Containsから完全一致に変更
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

// TestExecuteCreateDirectory はディレクトリの作成をテストします。
func TestExecuteCreateDirectory(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testDirName := "test_directory/"
	output, err := executeAndCaptureOutput(t, []string{"makit", testDirName})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if _, err := os.Stat(testDirName); os.IsNotExist(err) {
		t.Errorf("Directory %s was not created", testDirName)
	}
	expectedOutput := "Created directory: " + testDirName + "\n" // ★改行を追加
	if output != expectedOutput { // ★strings.Containsから完全一致に変更
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

// TestExecuteCreateNestedDirectory はネストされたディレクトリの作成をテストします。
func TestExecuteCreateNestedDirectory(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testPath := "parent/child/grandchild/"
	output, err := executeAndCaptureOutput(t, []string{"makit", testPath})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("Nested directory %s was not created", testPath)
	}
	expectedOutput := "Created directory: " + testPath + "\n" // ★改行を追加
	if !strings.Contains(output, expectedOutput) { // MkdirAllのverbose出力によってはContainsの方が良い
		t.Errorf("Expected output %q not found in %q", expectedOutput, output)
	}
}

// TestExecuteCreateFileInNewDirectory は新しいディレクトリ内にファイルを作成するテストです。
func TestExecuteCreateFileInNewDirectory(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testPath := "new_dir/new_file.txt"
	output, err := executeAndCaptureOutput(t, []string{"makit", testPath})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("File %s was not created in new directory", testPath)
	}
	expectedFileCreationOutput := "Created file: " + testPath + "\n" // ★改行を追加
	// MkdirAll の出力があるかもしれないので Contains を使用
	if !strings.Contains(output, expectedFileCreationOutput) {
		t.Errorf("Expected output %q not found in %q", expectedFileCreationOutput, output)
	}
}


// TestExecuteNoCreateExistingFile は -c オプションと既存ファイルをテストします。
func TestExecuteNoCreateExistingFile(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testFileName := "existing.txt"
	// 事前にファイルを作成
	err := ioutil.WriteFile(testFileName, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("Failed to create pre-existing file: %v", err)
	}

	output, cmdErr := executeAndCaptureOutput(t, []string{"makit", "-c", testFileName})
	if cmdErr != nil {
		t.Fatalf("Execute() failed: %v", cmdErr)
	}

	expectedOutput := "Exists: " + testFileName + "\n" // ★改行を追加
	if output != expectedOutput { // ★完全一致に変更
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
	cleanup := setupTestEnv(t)
	defer cleanup()

	testFileName := "non_existing.txt"
	output, cmdErr := executeAndCaptureOutput(t, []string{"makit", "-c", testFileName})
	if cmdErr != nil {
		t.Fatalf("Execute() failed: %v", cmdErr)
	}

	if _, err := os.Stat(testFileName); !os.IsNotExist(err) {
		t.Errorf("File %s was unexpectedly created", testFileName)
	}
	expectedOutput := "Skipped (not created): " + testFileName + "\n" // ★改行を追加
	if output != expectedOutput { // ★完全一致に変更
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

// TestExecuteInvalidMode は無効なモード指定をテストします。
func TestExecuteInvalidMode(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	output, err := executeAndCaptureOutput(t, []string{"makit", "-m", "invalid_mode", "dummy.txt"})
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
	cleanup := setupTestEnv(t)
	defer cleanup()

	testFileName := "mode_test.txt"
	modeStr := "755"
	expectedPerm := os.FileMode(0755)

	_, err := executeAndCaptureOutput(t, []string{"makit", "-m", modeStr, testFileName})
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
	cleanup := setupTestEnv(t)
	defer cleanup()

	output, err := executeAndCaptureOutput(t, []string{"makit", "-d", "invalid_timestamp", "dummy.txt"})
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
	cleanup := setupTestEnv(t)
	defer cleanup()

	testFileName := "timestamp_test.txt"
	timestampStr := "202401011030"
	expectedTime, _ := time.Parse("200601021504", timestampStr)

	_, err := executeAndCaptureOutput(t, []string{"makit", "-d", timestampStr, testFileName})
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
	cleanup := setupTestEnv(t)
	defer cleanup()

	testFileName := "verbose_file.txt"
	output, err := executeAndCaptureOutput(t, []string{"makit", "-v", testFileName})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// 詳細出力が含まれていることを確認
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
	cleanup := setupTestEnv(t)
	defer cleanup()

	args := []string{"file1.txt", "dir1/", "dir2/file2.txt"}
	output, err := executeAndCaptureOutput(t, append([]string{"makit"}, args...))
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
	cleanup := setupTestEnv(t)
	defer cleanup()

	// os.Argsをnilにリセットし、引数を空にする
	// CobraのMinimumNArgs(1)によりエラーとなることを期待
	os.Args = []string{"makit"} // Execute() が os.Args を直接読むため、これを設定

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	
	err := rootCmd.Execute() // Execute() を呼び出す

	// 出力を元に戻す
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	if err == nil {
		t.Errorf("Expected an error for no arguments, but got none")
	}
	// エラーメッセージの確認（Cobraのデフォルトエラーメッセージを期待）
	expectedErrorOutput := "Error: accepts at least 1 arg(s), received 0"
	if !strings.Contains(buf.String(), expectedErrorOutput) {
		t.Errorf("Expected error message %q not found in %q", expectedErrorOutput, buf.String())
	}
}