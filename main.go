package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"map/app/commands"
	"map/app/services"
	"map/app/views"
)

func main() {
	//@ai:think intent=程式進入點：先登錄所有 schema 再依旗標分派 GUI / CLI 兩種執行模式 model=claude-opus-4-8
	// 啟動時把所有 schema 登錄到 schema.Global()（供 hyp context / lint / contract）。
	commands.RegisterSchemas()
	views.RegisterSchemas()

	guiMode := flag.Bool("windows", false, "開啟圖形視窗 (GUI) 模式")
	// -i / -o 的旗標註冊由 commands.MapCommand 的 schema 驅動。
	vals := commands.BindFlags(flag.CommandLine)
	flag.Usage = usage
	flag.Parse()

	// --windows：開啟視窗模式，其餘一律走命令列模式。
	if *guiMode {
		runGUI()
		return
	}
	runCLI(vals)
}

// runGUI 啟動 Fyne 桌面視窗，顯示主畫面。
func runGUI() {
	//@ai:think intent=視窗模式：建立 Fyne 視窗並顯示主畫面，供非技術使用者以 GUI 操作 model=claude-opus-4-8
	a := app.New()
	w := a.NewWindow("Map")
	w.Resize(fyne.NewSize(800, 600))
	w.SetContent(views.MainView(w))
	w.ShowAndRun()
}

// runCLI 為命令列模式：依 schema 檢查 -i / -o，呼叫 service 產生 map。
func runCLI(vals map[string]*string) {
	//@ai:think intent=命令列模式：依 schema 驗證必填旗標後呼叫 service 產生 map model=claude-opus-4-8
	//@ai:think special=缺參數回 exit 2、轉換失敗回 exit 1，方便腳本判斷狀態 model=claude-opus-4-8
	if missing := commands.MissingRequired(vals); len(missing) > 0 {
		fmt.Fprintln(os.Stderr, "缺少必填參數:", strings.Join(missing, ", "))
		flag.Usage()
		os.Exit(2)
	}
	res, err := services.BuildMap(*vals["i"], *vals["o"])
	if err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
	fmt.Printf("✓ 掃描 %d 個套件、%d 個 .go 檔，已輸出 %d 篇筆記到 %s\n",
		res.Packages, res.GoFiles, res.Notes, res.OutputDir)
}

// usage 印出命令列說明；-i / -o 由 commands.MapCommand 的 schema 自動產生。
func usage() {
	//@ai:think intent=由 schema 的 Params 動態產生說明，讓旗標與說明文字維持單一事實來源 model=claude-opus-4-8
	fmt.Fprintf(os.Stderr, "GoMap — %s\n\n", commands.MapCommand.Summary)
	fmt.Fprint(os.Stderr, "用法:\n")
	fmt.Fprint(os.Stderr, "  map --windows                      開啟圖形視窗 (GUI)\n")
	fmt.Fprint(os.Stderr, "  map -i <go目錄> -o <obsidian目錄>   命令列模式，掃描並輸出\n\n")
	fmt.Fprint(os.Stderr, "參數:\n")
	fmt.Fprint(os.Stderr, "  --windows   開啟視窗模式\n")
	for _, p := range commands.MapCommand.Params {
		req := "選填"
		if p.Required {
			req = "必填"
		}
		fmt.Fprintf(os.Stderr, "  %s %s   %s (%s)\n", commands.FlagToken(p.Name), p.Type, p.Desc, req)
	}
}
