package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/ZxwyWebSite/ovi-share/config"
	"github.com/ZxwyWebSite/ovi-share/server"
)

func init() {
	os.Stdout.WriteString("== OVI-Share v0.0.3 (r260510) ==\n\nThanks: SpencerWooo/OneDrive-Vercel-Index\nMade with ❤ by Zxwy & OvList Contributors.\n\n")
}

func main() {
	cfg, err := config.Load(`data/config.json`)
	if err != nil {
		fmt.Printf("初始化配置文件：%s\n", err)
		return
	}

	os.Stdout.WriteString("init fs...\n\n")

	root, site, err := cfg.Build(context.Background())
	if err != nil {
		fmt.Printf("初始化文件系统：%s\n", err)
		return
	}

	r := server.Router(root, site, cfg)

	fmt.Printf("\nServer started on %s\n", cfg.Serv.Listen)

	if err := http.ListenAndServe(cfg.Serv.Listen, r); err != nil {
		fmt.Printf("服务器监听失败：%s\n", err)
	}
}
