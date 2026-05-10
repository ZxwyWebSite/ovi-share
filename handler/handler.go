package handler

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/ZxwyWebSite/ovi-share/config"
	"github.com/ZxwyWebSite/ovi-share/pkg/util"
	"github.com/ZxwyWebSite/ovi-share/pkg/vfs"
	"github.com/ZxwyWebSite/ovi-share/provider/share"
	"github.com/coocood/freecache"
	"golang.org/x/sync/singleflight"
)

type Handler struct {
	Root  vfs.Provider
	Site  map[string]vfs.Provider
	Odpt  []config.CfgOdpt
	Cache *freecache.Cache
	SF    singleflight.Group
}

func New(root vfs.Provider, site map[string]vfs.Provider, odpt []config.CfgOdpt, cache int) *Handler {
	// 初始化密码与哈希缓存
	sort.Slice(odpt, func(i, j int) bool {
		return odpt[i].Prefix > odpt[j].Prefix
	})
	for i, v := range odpt {
		hash := sha256.Sum256(util.StringToBytes(v.Password))
		odpt[i].Cache = util.BytesToString(util.HexEncode(hash[:]))
	}
	return &Handler{
		Root:  root,
		Site:  site,
		Odpt:  odpt,
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
