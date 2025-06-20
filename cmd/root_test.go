package cmd

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMakit_CreateFile(t *testing.T) {
	tempFile := "test_output.txt"
	defer os.Remove(tempFile) // 後始末

	cmd := exec.Command("go", "run", "../main.go", tempFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, string(output))
	}

	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Errorf("expected file to be created: %s", tempFile)
	}
}

func TestMakit_NoArgs(t *testing.T) {
	cmd := exec.Command("go", "run", "../main.go")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error due to missing args, but got none")
	}

	// エラーメッセージに usage 含まれることを確認
	if !strings.Contains(string(output), "Usage:") {
		t.Errorf("expected usage message in output, got: %s", string(output))
	}
}
