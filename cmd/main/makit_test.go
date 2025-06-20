package main

/*
// Example_makit は、goMain() 関数が削除されたため、コメントアウトします。
func Example_makit() {
    goMain([]string{"makit!"})
    // Output:
    // Welcome to makit!
}
*/

/*
// TestHello は、hello() 関数が削除されたため、コメントアウトします。
func TestHello(t *testing.T) {
    got := hello()
    want := "Welcome to makit!"
    if got != want {
        t.Errorf("hello() = %q, want %q", got, want)
    }
}
*/

// 今後、もし makit CLIツールの機能（ファイル作成など）をテストしたい場合、
// cmd/root.go の Execute() を呼び出し、その動作を検証するような、
// 新しいテストをここに記述する必要があります。
// 例: コマンドライン引数をシミュレートし、ファイルが作成されたか確認する、など。