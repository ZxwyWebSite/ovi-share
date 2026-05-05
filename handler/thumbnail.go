package handler

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/ZxwyWebSite/ovi-share/pkg/util"
	"github.com/ZxwyWebSite/ovi-share/pkg/vfs"
)

var ErrThumbNotFound = errors.New(`thumb not found`)

// 获取缩略图
func (a *Handler) Thumbnail(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	qsize := q.Get(`size`)

	var i int
	switch qsize {
	case `large`:
		i = 2
	case `medium`:
		i = 1
	case `small`:
		i = 0
	default:
		Error(w, `Invalid size`, http.StatusBadRequest)
		return
	}

	qpath := q.Get(`path`)

	// 目录穿越检测
	if containsDotDot(qpath) {
		Error(w, `invalid URL path`, http.StatusBadRequest)
		return
	}
	qpath = path.Clean(qpath)

	// 站点
	var tk []byte // thumb
	ti := string(byte('0' + i))
	fs, vhost := a.Site[r.Host]
	if fs == nil {
		fs = a.Root
		vhost = false
		tk = util.ConcatB(`t:`, qpath, `:`, ti)
	} else {
		tk = util.ConcatB(r.Host, `:t:`, qpath, `:`, ti)
	}

	val, exp, err := a.Cache.GetWithExpiration(tk)
	if err == nil {
		w.Header()[hkCC] = []string{hvCC + strconv.FormatInt(int64(exp)-time.Now().Unix(), 10)}
		http.Redirect(w, r, util.BytesToString(val), http.StatusFound)
		return
	}

	if fs, ok := fs.(vfs.Thumb); ok {
		// 三种大小一起获取，忽略尾部 size
		v, err, _ := a.SF.Do(util.BytesToString(tk[:len(tk)-2]), func() (any, error) {
			os.Stdout.Write(util.ConcatB(`[OVI] thumb: `, qpath, "\n"))
			thumbs, err := fs.Thumb(r.Context(), qpath)
			if err != nil {
				return nil, err
			}
			if len(thumbs) == 3 {
				var stk, mtk, ltk []byte
				if vhost {
					stk = util.ConcatB(r.Host, `:t:`, qpath, `:0`)
					mtk = util.ConcatB(r.Host, `:t:`, qpath, `:1`)
					ltk = util.ConcatB(r.Host, `:t:`, qpath, `:2`)
				} else {
					stk = util.ConcatB(`t:`, qpath, `:0`)
					mtk = util.ConcatB(`t:`, qpath, `:1`)
					ltk = util.ConcatB(`t:`, qpath, `:2`)
				}
				a.Cache.Set(stk, util.StringToBytes(thumbs[0]), maxAge)
				a.Cache.Set(mtk, util.StringToBytes(thumbs[1]), maxAge)
				a.Cache.Set(ltk, util.StringToBytes(thumbs[2]), maxAge)
			} else {
				return nil, ErrThumbNotFound
			}
			return thumbs, nil
		})
		if err != nil {
			Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header()[hkCC] = CC
		http.Redirect(w, r, v.([]string)[i], http.StatusFound)
		return
	}
	Error(w, `vfs: thumb not support`, http.StatusNotImplemented)
}
