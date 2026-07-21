package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSearchCmd(t *testing.T) {
	// 1. テスト用のファイルを作成（t.TempDir() はテスト終了時に自動削除されます）
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.txt")
	fileContent := "Hello World\nhello Go\nHELLO Cobra"

	if err := os.WriteFile(testFile, []byte(fileContent), 0o644); err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// 2. テーブル駆動テストケースの定義
	tests := []struct {
		name         string
		args         []string
		wantCountStr string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "正常系: 大文字小文字を区別して検索 (1回一致)",
			args:         []string{"search", "-f", testFile, "-t", "hello"},
			wantCountStr: "[hello]の出現回数: 1回",
			wantErr:      false,
		},
		{
			name:         "正常系: 大文字小文字を無視して検索 (-i 指定で 3回一致)",
			args:         []string{"search", "-f", testFile, "-t", "hello", "-i"},
			wantCountStr: "[hello]の出現回数: 3回",
			wantErr:      false,
		},
		{
			name:        "異常系: 存在しないファイルパス",
			args:        []string{"search", "-f", filepath.Join(tmpDir, "not_exist.txt"), "-t", "hello"},
			wantErr:     true,
			errContains: "searching your file is failed",
		},
		{
			name:    "異常系: 必須フラグ (--target) なし",
			args:    []string{"search", "-f", testFile},
			wantErr: true, // Cobraが自動でエラーを返します
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストごとにグローバル変数と Cobra のフラグ状態をリセット
			filePath = ""
			word = ""
			ignoreCase = false
			if f := searchCmd.Flags().Lookup("file"); f != nil {
				f.Changed = false
			}
			if f := searchCmd.Flags().Lookup("target"); f != nil {
				f.Changed = false
			}
			if f := searchCmd.Flags().Lookup("ignore-case"); f != nil {
				f.Changed = false
			}

			// 標準出力・エラー出力を受け取るバッファを作成
			outBuf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)

			// Cobraコマンドにバッファと引数をセット
			rootCmd.SetOut(outBuf)
			rootCmd.SetErr(errBuf)
			rootCmd.SetArgs(tt.args)

			// コマンド実行
			err := rootCmd.Execute()

			// エラーの検証
			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("期待するエラー文字列: %q, 実際のエラー: %q", tt.errContains, err.Error())
				}
			}

			// 正常系の出力メッセージの検証
			if !tt.wantErr && tt.wantCountStr != "" {
				output := outBuf.String()
				if !strings.Contains(output, tt.wantCountStr) {
					t.Errorf("出力に期待する文字列が含まれていません。\n期待値: %q\n実際の出力:\n%s", tt.wantCountStr, output)
				}
			}
		})
	}
}
