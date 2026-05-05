package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ZxwyWebSite/ovi-share/pkg/vfs"
	"github.com/ZxwyWebSite/ovi-share/provider/share"
	"github.com/coocood/freecache"
	"golang.org/x/sync/singleflight"
)

type Handler struct {
	Root  vfs.Provider
	Site  map[string]vfs.Provider
	Cache *freecache.Cache
	SF    singleflight.Group
}

func New(root vfs.Provider, site map[string]vfs.Provider, cache int) *Handler {
	return &Handler{
		Root:  root,
		Site:  site,
		Cache: freecache.NewCache(cache * 1024 * 1024),
	}
}

func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for ent := range strings.FieldsFuncSeq(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func isSlashRune(r rune) bool { return r == '/' || r == '\\' }

type AppError struct {
	Error string `json:"error"`
}

// 返回 JSON 格式错误
func Error(w http.ResponseWriter, err string, code int) {
	w.Header()[`Content-Type`] = ctJSON
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(AppError{Error: err})
}

type AppFile struct {
	File *share.DriveItem `json:"file"`
}

type AppFolder struct {
	Folder *share.DriveChildren `json:"folder"`

	Next string `json:"next,omitempty"`
}

func Json(w http.ResponseWriter, v any, code int) {
	w.Header()[`Content-Type`] = ctJSON
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func JsonBlob(w http.ResponseWriter, v []byte, code int) {
	w.Header()[`Content-Type`] = ctJSON
	w.WriteHeader(code)
	w.Write(v)
}
