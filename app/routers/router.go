// Package routers 提供桌面 GUI 的「畫面路由器」(view router)。
//
// 採 HypGo schema-first 規則：每個畫面先以 schema.Route 宣告（ViewRoutes），
// 註冊到 schema.Global()，路由器再依宣告「驗證」綁定與導航——
//   - 只能綁定 (Register) 已宣告的畫面；
//   - 導航 (NavigateWith) 帶入的資料型別必須符合該畫面宣告的 Input。
// 與 cli.go 的 MapCommand 同一精神（宣告即合約，供 hyp context / manifest 取用）。
//
// 典型流程 (GoMap)：
//
//	home  → 選擇 Go 專案目錄
//	scan  → 掃描中 / 進度
//	map   → 預覽並匯出 Obsidian map
//	settings → 設定
package routers

import (
	"fmt"
	"reflect"

	"map/app/models"
	"map/app/views"

	"fyne.io/fyne/v2"
	"github.com/maoxiaoyue/hypgo/pkg/schema"
)

// GUIProtocol 是桌面畫面導航在 HypGo schema 中的協議名稱。
const GUIProtocol = "view"

// 畫面名稱常數，即 schema.Route.Command。註冊與導航時請使用這些常數。
const (
	RouteHome     = "home"     // 選擇要建立 Obsidian map 的 Go 專案目錄
	RouteScanning = "scanning" // 掃描專案、解析 package/import 關係
	RouteMap      = "map"      // 預覽產生的 map、設定輸出位置並匯出
	RouteSettings = "settings" // 應用程式設定
)

// 畫面導航帶入的資料契約定義於 models（即 schema 的 Input 型別）：
// scanning → models.ScanInput；map → models.MapResult（直接沿用掃描結果）。

// ViewRoutes 是所有 GUI 畫面的 schema 宣告（schema-first 的單一事實來源）。
var ViewRoutes = []schema.Route{
	{Protocol: GUIProtocol, Command: RouteHome, Summary: "選擇要建立 Obsidian map 的 Go 專案目錄"},
	{Protocol: GUIProtocol, Command: RouteScanning, Summary: "掃描專案、解析 package/import 關係",
		Input: models.ScanInput{}, InputName: schema.TypeName(models.ScanInput{})},
	{Protocol: GUIProtocol, Command: RouteMap, Summary: "預覽產生的 map、設定輸出位置並匯出",
		Input: models.MapResult{}, InputName: schema.TypeName(models.MapResult{})},
	{Protocol: GUIProtocol, Command: RouteSettings, Summary: "應用程式設定"},
}

// init 於套件載入時將所有畫面 schema 註冊到全域註冊表。
func init() {
	for _, rt := range ViewRoutes {
		schema.Global().Register(rt)
	}
}

// IsDeclaredView 回報 name 是否為 ViewRoutes schema 中宣告過的畫面。
func IsDeclaredView(name string) bool {
	_, ok := schema.Global().GetByKey(GUIProtocol + "|" + name)
	return ok
}

// ViewFactory 依需求建立某個畫面的內容。
//
// 每次導航到該路由時都會被呼叫一次（不會快取），因此回傳的畫面永遠是最新狀態。
// factory 透過傳入的 *Router 觸發後續導航（例如按鈕點擊後 r.Navigate(...)），
// 並可用 r.Window() 取得視窗、r.Data() 取得本次導航帶入的參數。
type ViewFactory func(r *Router) fyne.CanvasObject

// navEntry 記錄一次導航的目標路由與其攜帶的資料，供返回 (Back) 時還原。
type navEntry struct {
	name string
	data any
}

// Router 管理單一視窗內的畫面切換。
//
// 注意：所有導航方法都會更新視窗內容，必須在 Fyne 的 UI goroutine 上呼叫。
// 若需從背景 goroutine（例如掃描完成後）切換畫面，請以 fyne.Do(func(){ r.Navigate(...) }) 包裹。
type Router struct {
	window  fyne.Window
	routes  map[string]ViewFactory
	history []navEntry
	current navEntry
}

// New 以指定視窗建立一個空的 Router。
func New(w fyne.Window) *Router {
	return &Router{
		window: w,
		routes: make(map[string]ViewFactory),
	}
}

// Register 將一個畫面名稱對應到它的 factory，回傳自身以便串接。
//
// 依 schema 規則：name 必須是 ViewRoutes 中宣告過的畫面，否則 panic（程式設定錯誤）。
// factory 不可為 nil；重複註冊會覆寫先前的設定。
func (r *Router) Register(name string, factory ViewFactory) *Router {
	if factory == nil {
		panic("routers: Register 的 factory 不可為 nil (route=" + name + ")")
	}
	if !IsDeclaredView(name) {
		panic("routers: 路由 " + name + " 未在 ViewRoutes schema 中宣告，無法註冊")
	}
	r.routes[name] = factory
	return r
}

// Window 回傳此路由器綁定的視窗。
func (r *Router) Window() fyne.Window { return r.window }

// Current 回傳目前顯示的路由名稱（尚未導航時為空字串）。
func (r *Router) Current() string { return r.current.name }

// Data 回傳本次導航帶入的資料，由畫面 factory 自行型別斷言取用。
//
// 例如：home 選好目錄後 NavigateWith(RouteScanning, models.ScanInput{Dir: dir})，
// scanning 畫面以 in, _ := r.Data().(models.ScanInput) 取得資料。
func (r *Router) Data() any { return r.current.data }

// Navigate 切換到指定路由（不帶資料）。
func (r *Router) Navigate(name string) error {
	return r.NavigateWith(name, nil)
}

// NavigateWith 切換到指定路由並附帶資料。
// 會先依 schema 驗證資料型別，再把目前畫面推入歷史堆疊，呼叫目標 factory 設定到視窗。
func (r *Router) NavigateWith(name string, data any) error {
	factory, ok := r.routes[name]
	if !ok {
		return fmt.Errorf("routers: 未註冊的路由 %q", name)
	}
	if err := validateInput(name, data); err != nil {
		return err
	}

	if r.current.name != "" {
		r.history = append(r.history, r.current)
	}
	r.current = navEntry{name: name, data: data}

	r.window.SetContent(factory(r))
	return nil
}

// validateInput 依 schema 規則檢查帶入資料是否符合該畫面宣告的 Input 契約。
func validateInput(name string, data any) error {
	route, ok := schema.Global().GetByKey(GUIProtocol + "|" + name)
	if !ok {
		return nil // 無 schema 宣告則不檢查
	}
	if route.Input == nil {
		if data != nil {
			return fmt.Errorf("routers: 畫面 %q 未宣告 Input，但收到 %T", name, data)
		}
		return nil
	}
	if data == nil {
		return nil // 允許先導航、稍後再帶資料
	}
	if want, got := reflect.TypeOf(route.Input), reflect.TypeOf(data); want != got {
		return fmt.Errorf("routers: 畫面 %q 期望 Input 型別 %s，但收到 %s", name, want, got)
	}
	return nil
}

// CanGoBack 回報歷史堆疊中是否還有可返回的畫面。
func (r *Router) CanGoBack() bool { return len(r.history) > 0 }

// Back 返回上一個畫面（連同當時帶入的資料一併還原）。
// 若已無上一頁則回傳錯誤。
func (r *Router) Back() error {
	if !r.CanGoBack() {
		return fmt.Errorf("routers: 已無可返回的畫面")
	}

	last := len(r.history) - 1
	prev := r.history[last]
	r.history = r.history[:last]
	r.current = prev

	r.window.SetContent(r.routes[prev.name](r))
	return nil
}

// Setup 建立 Router、註冊所有畫面，並導航到首頁。
// main.go 只需呼叫此函式即可取得已就緒的路由器。
//
// 目前僅 home 對應到既有的 views.MainView；scanning / map / settings 已在
// ViewRoutes 宣告，待對應畫面 (app/views/*) 建立後，於此處接上即可：
//
//	r.Register(RouteScanning, func(r *Router) fyne.CanvasObject {
//	    in := r.Data().(models.ScanInput)
//	    return views.ScanningView(r.Window(), in.Dir)
//	})
//	r.Register(RouteMap, func(r *Router) fyne.CanvasObject { ... })
//	r.Register(RouteSettings, func(r *Router) fyne.CanvasObject { ... })
func Setup(w fyne.Window) *Router {
	r := New(w)

	r.Register(RouteHome, func(r *Router) fyne.CanvasObject {
		return views.MainView(r.Window())
	})

	// 進入點：首頁。忽略錯誤，因為 RouteHome 必定已註冊。
	_ = r.Navigate(RouteHome)
	return r
}
