package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZxwyWebSite/ovi-share/handler"
	"github.com/ZxwyWebSite/ovi-share/pkg/util"
	"github.com/ZxwyWebSite/ovi-share/pkg/vfs"
)

// 单页应用程序处理器
func spaHandler(staticDir string, indexFile string) http.HandlerFunc {
	indexFile = filepath.Join(staticDir, indexFile)
	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(staticDir, r.URL.Path)

		// 判断文件是否存在
		_, err := os.Stat(path)
		if os.IsNotExist(err) || strings.HasSuffix(r.URL.Path, "/") {
			// fallback 到 index.html
			http.ServeFile(w, r, indexFile)
			return
		}

		// 静态文件
		http.FileServer(http.Dir(staticDir)).ServeHTTP(w, r)
	}
}

// 初始化路由
func Router(fs vfs.Provider, cache int, static string) http.Handler {
	h := handler.New(fs, cache)

	mux := http.NewServeMux()

	/*mux.HandleFunc(`/api/`, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodPost {
			h.Index(w, r)
			return
		}
		http.NotFound(w, r)
	})*/

	mux.HandleFunc(`GET /api/`, h.Index)
	mux.HandleFunc(`GET /api/name/`, h.Raw)
	mux.HandleFunc(`GET /api/item/`, h.Item)
	mux.HandleFunc(`GET /api/raw/`, h.Raw)
	mux.HandleFunc(`GET /api/search/`, h.Search)
	mux.HandleFunc(`GET /api/thumbnail/`, h.Thumbnail)

	mux.HandleFunc(`GET /`, spaHandler(static, `index.html`))

	// Logger
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		s := &statusRecorder{w, http.StatusOK}

		mux.ServeHTTP(s, r)

		var uri string
		if r.URL.RawQuery != `` {
			uri = util.Concat(r.URL.Path, `?`, r.URL.RawQuery)
		} else {
			uri = r.URL.Path
		}

		fmt.Fprintf(os.Stdout,
			"[OVI] %s | %3d | %13v | %15s | %-7s  \"%s\"\n",
			start.Format(time.DateTime),
			s.Status,
			time.Since(start),
			realIP(r),
			r.Method,
			uri,
		)
	})
}

// 记录响应码
type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.Status = statusCode
}

// 解析客户端 IP（ECHO）
func realIP(r *http.Request) string {
	// Fall back to legacy behavior
	if h := r.Header[`X-Forwarded-For`]; len(h) != 0 && h[0] != "" {
		i := strings.IndexAny(h[0], ",")
		if i > 0 {
			xffip := strings.TrimSpace(h[0][:i])
			xffip = strings.TrimPrefix(xffip, "[")
			xffip = strings.TrimSuffix(xffip, "]")
			return xffip
		}
		return h[0]
	}
	if h := r.Header[`X-Real-Ip`]; len(h) != 0 && h[0] != "" {
		ip := strings.TrimPrefix(h[0], "[")
		ip = strings.TrimSuffix(ip, "]")
		return ip
	}
	ra, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ra
}
