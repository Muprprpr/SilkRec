package main

import (
	"embed"
	"net/http"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// 确保output目录存在
	os.MkdirAll("output", 0755)

	// 创建自定义的HTTP处理器来服务output目录
	// 这样前端可以通过 http://localhost:port/output/xxx.webm 访问视频文件
	mux := http.NewServeMux()

	// 添加output目录的静态文件服务
	mux.Handle("/output/", http.StripPrefix("/output/", http.FileServer(http.Dir("output"))))

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "SilkRec",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: mux, // 使用自定义的HTTP处理器
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
