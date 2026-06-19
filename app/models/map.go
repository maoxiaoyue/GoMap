// Package models 定義 GoMap 的資料模型（MVC 的 M）：
// 在 service、router(controller) 與 view 之間傳遞的 DTO / 型別契約。
//
// 這些型別同時作為 HypGo schema 的 Input/Output 契約——json tag 即欄位名，
// 非指標且無 omitempty 者，schema.FieldsOf 會判定為必填。
package models

// CLIInput 是命令列模式的輸入契約（-i / -o）。
type CLIInput struct {
	InputDir  string `json:"i"` // 輸入：要掃描的 Go 專案目錄
	OutputDir string `json:"o"` // 輸出：產生 Obsidian map 的目錄 (vault)
}

// MapResult 是一次轉換的結果摘要。
//
// 由 services.BuildMap 產出，並作為 CLI 命令 (commands.MapCommand) 的 schema Output 契約。
type MapResult struct {
	Packages  int    `json:"packages"`   // 掃描到的套件數
	GoFiles   int    `json:"go_files"`   // 納入的 .go 檔數（不含 _test.go）
	Functions int    `json:"functions"`  // 寫出的函式說明節點數（帶 @ai 標註者）
	Notes     int    `json:"notes"`      // 寫出的筆記數（套件 + 索引 + HypGo Context + 函式節點）
	OutputDir string `json:"output_dir"` // 實際輸出目錄
}
