package routers

// cli.go 以 HypGo 的 schema-first 方式宣告命令列模式的契約（-i / -o）。
//
// 由人類用資料宣告命令、旗標、型別與是否必填，工具與 AI（hyp context /
// manifest / contract）不必讀 main() 的 flag 接線即可理解這個 CLI 命令。
// schema 會在 import 本套件時自動註冊到 schema.Global()。

import (
	"flag"
	"strings"

	"map/app/models"

	"github.com/maoxiaoyue/hypgo/pkg/schema"
)

// MapCommand 是 GoMap 命令列模式的 schema 定義（Protocol "cli"）。
//
// Input/Output 引用 models 的型別契約（models.CLIInput / models.MapResult）；
// Params 則以旗標層級描述 -i / -o，讓 schema 對 CLI 工具自我描述。
var MapCommand = schema.Route{
	Protocol:   "cli",
	Command:    "map",
	Summary:    "將指定的 Go 專案目錄轉成 Obsidian map",
	Tags:       []string{"gomap"},
	Input:      models.CLIInput{},
	Output:     models.MapResult{},
	InputName:  schema.TypeName(models.CLIInput{}),
	OutputName: schema.TypeName(models.MapResult{}),
	Params: []schema.ParamSchema{
		{Name: "i", In: "flag", Required: true, Type: "string", Desc: "輸入：要掃描的 Go 專案目錄"},
		{Name: "o", In: "flag", Required: true, Type: "string", Desc: "輸出：產生 Obsidian map 的目錄 (vault)"},
	},
}

// init 於套件載入時，將 CLI schema 註冊到全域註冊表，
// 供 hyp context / manifest / contract 取用。
func init() {
	schema.Global().Register(MapCommand)
}

// FlagToken 將旗標名加上前綴：單字元用 "-x"、多字元用 "--xx"。
func FlagToken(name string) string {
	if len(name) == 1 {
		return "-" + name
	}
	return "--" + name
}

// BindFlags 依 MapCommand 的 Params 在 fs 上註冊旗標（schema 為單一事實來源），
// 回傳「旗標名 → 值指標」對照表，供解析後取值。
func BindFlags(fs *flag.FlagSet) map[string]*string {
	vals := make(map[string]*string, len(MapCommand.Params))
	for _, p := range MapCommand.Params {
		vals[p.Name] = fs.String(p.Name, "", p.Desc)
	}
	return vals
}

// MissingRequired 依 schema 檢查必填旗標，回傳尚未提供值者（含前綴），皆備齊則為空。
func MissingRequired(vals map[string]*string) []string {
	var missing []string
	for _, p := range MapCommand.Params {
		if !p.Required {
			continue
		}
		if v, ok := vals[p.Name]; !ok || strings.TrimSpace(*v) == "" {
			missing = append(missing, FlagToken(p.Name))
		}
	}
	return missing
}
