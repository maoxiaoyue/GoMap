package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"map/app/routers"
	"map/app/services"
)

func main() {
	guiMode := flag.Bool("windows", false, "開啟圖形視窗 (GUI) 模式")
	// -i / -o 的旗標註冊由 routers.MapCommand 的 schema 驅動。
	vals := routers.BindFlags(flag.CommandLine)
	flag.Usage = usage
	flag.Parse()

	// --windows：開啟視窗模式，其餘一律走命令列模式。
	if *guiMode {
		runGUI()
		return
	}

	// 依 schema 檢查必填旗標。
	if missing := routers.MissingRequired(vals); len(missing) > 0 {
		fmt.Fprintln(os.Stderr, "缺少必填參數:", strings.Join(missing, ", "))
		flag.Usage()
		os.Exit(2)
	}

	res, err := services.BuildMap(*vals["i"], *vals["o"])
	if err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}

	// res 為 models.MapResult（schema 的輸出契約）。
	fmt.Printf("✓ 掃描 %d 個套件、%d 個 .go 檔，已輸出 %d 篇筆記到 %s\n",
		res.Packages, res.GoFiles, res.Notes, res.OutputDir)
}

// runGUI 啟動 Fyne 桌面視窗，由畫面路由器接管畫面切換。
func runGUI() {
	a := app.New()
	w := a.NewWindow("Map")
	w.Resize(fyne.NewSize(800, 600))

	routers.Setup(w)

	w.ShowAndRun()
}

// usage 印出命令列說明；-i / -o 由 routers.MapCommand 的 schema 自動產生。
func usage() {
	fmt.Fprintf(os.Stderr, "GoMap — %s\n\n", routers.MapCommand.Summary)
	fmt.Fprint(os.Stderr, "用法:\n")
	fmt.Fprint(os.Stderr, "  map --windows                      開啟圖形視窗 (GUI)\n")
	fmt.Fprint(os.Stderr, "  map -i <go目錄> -o <obsidian目錄>   命令列模式，掃描並輸出\n\n")
	fmt.Fprint(os.Stderr, "參數:\n")
	fmt.Fprint(os.Stderr, "  --windows   開啟視窗模式\n")
	for _, p := range routers.MapCommand.Params {
		req := "選填"
		if p.Required {
			req = "必填"
		}
		fmt.Fprintf(os.Stderr, "  %s %s   %s (%s)\n", routers.FlagToken(p.Name), p.Type, p.Desc, req)
	}
}
