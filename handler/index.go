package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ZxwyWebSite/ovi-share/pkg/util"
	"github.com/ZxwyWebSite/ovi-share/pkg/vfs"
	"github.com/ZxwyWebSite/ovi-share/provider/share"
)

const (
	hkCC = `Cache-Control`
	hvCC = `max-age=`

	maxAge = 3500
)

var (
	CC = []string{`max-age=3500`}

	ctJSON = []string{`application/json`}
)

// 文件与目录综合处理
func (a *Handler) Index(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	qpath := q.Get(`path`)

	// 目录穿越检测
	if containsDotDot(qpath) {
		Error(w, `invalid URL path`, http.StatusBadRequest)
		return
	}
	qpath = path.Clean(qpath)

	c := r.Context()

	raw := c.Value(`raw`) == true || q.Has(`raw`)

	var lk, ik []byte // link, info
	// 站点
	fs, vhost := a.Site[r.Host]
	if fs == nil {
		for i, l := 0, len(a.Odpt); i < l; i++ {
			if strings.HasPrefix(qpath, a.Odpt[i].Prefix) {
				odpt := q.Get(`odpt`)
				if odpt == `` {
					h := r.Header[`Od-Protected-Token`]
					if len(h) != 0 {
						odpt = h[0]
					}
				}
				w.Header()[`Vary`] = append(w.Header()[`Vary`], `Od-Protected-Token`)
				if odpt != a.Odpt[i].Cache {
					Error(w, `protected`, http.StatusUnauthorized)
					return
				}
				break
			}
		}
		fs = a.Root
		vhost = false
		lk = util.ConcatB(`l:`, qpath)
		ik = util.ConcatB(`i:`, qpath)
	} else {
		lk = util.ConcatB(r.Host, `:l:`, qpath)
		ik = util.ConcatB(r.Host, `:i:`, qpath)
	}

	// 目录暂时不缓存，所以用不到，只有文件才会命中缓存
	/*next := q.Get(`next`)
	var ik []byte
	if raw || next == `` {
		ik = util.ConcatB(`i:`, qpath) // info
	} else {
		ik = util.ConcatB(`i:`, qpath, `:`, next)
	}*/
	if raw {
		val, exp, err := a.Cache.GetWithExpiration(lk)
		if err == nil {
			w.Header()[hkCC] = []string{hvCC + strconv.FormatInt(int64(exp)-time.Now().Unix(), 10)}
			http.Redirect(w, r, util.BytesToString(val), http.StatusFound)
			return
		}
	} else {
		val, exp, err := a.Cache.GetWithExpiration(ik)
		if err == nil {
			w.Header()[hkCC] = []string{hvCC + strconv.FormatInt(int64(exp)-time.Now().Unix(), 10)}
			JsonBlob(w, val, http.StatusOK)
			return
		}
	}

	type ictx struct {
		n vfs.Node
		b []byte
		l string
		d *share.DriveChildren
	}
	v, err, _ := a.SF.Do(util.BytesToString(ik), func() (any, error) {
		os.Stdout.Write(util.ConcatB(`[OVI] open: `, qpath, "\n"))
		f, err := fs.Open(c, qpath)
		if err != nil {
			return nil, err
		}
		var ic = ictx{n: f}
		// 缓存详情与链接
		if p, ok := f.(share.Item); ok {
			if it := p.GetItem(); it.ContentDownloadURL != `` {
				// 文件
				d, _ := json.Marshal(AppFile{it})
				a.Cache.Set(ik, d, maxAge)
				a.Cache.Set(lk, util.StringToBytes(it.ContentDownloadURL), maxAge)
				ic.b = d
				ic.l = it.ContentDownloadURL
			} else {
				// 目录
				its, err := p.ListItem(c, ``, q.Get(`next`))
				if err != nil {
					return nil, err
				}
				l := len(its.Value)
				// 缓存子项目
				for i := 0; i < l; i++ {
					if itm := its.Value[i]; itm.ContentDownloadURL != `` {
						d, _ := json.Marshal(AppFile{&itm})
						itp := path.Join(qpath, itm.Name)
						var ikp, lkp []byte
						if vhost {
							ikp = util.ConcatB(r.Host, `:i:`, itp)
							lkp = util.ConcatB(r.Host, `:l:`, itp)
						} else {
							ikp = util.ConcatB(`i:`, itp)
							lkp = util.ConcatB(`l:`, itp)
						}
						a.Cache.Set(ikp, d, maxAge)
						a.Cache.Set(lkp, util.StringToBytes(itm.ContentDownloadURL), maxAge)
					}
				}
				// 提取 skiptoken
				if i := strings.LastIndexByte(its.OdataNextLink, '='); i != -1 {
					its.OdataNextLink = its.OdataNextLink[i+1:]
				}
				// 流式输出
				ic.d = its
				// 避免大目录 OOM
				/*if l <= 10 {
					d, _ := json.Marshal(AppFolder{share.DriveChildren{Value: its}})
					err = a.Cache.Set(ik, d, maxAge)
					if err != nil {
						println(err.Error())
					}
					ic.b = d
				}*/
			}
		} else if f.IsDir() {
			if d, ok := f.(vfs.Dir); ok {
				// 兼容 Mount 根目录
				ns, err := d.List(c, ``)
				if err != nil {
					return nil, err
				}
				l := len(ns)
				o := make([]share.DriveItem, l)
				for i := 0; i < l; i++ {
					o[i].Name = ns[i].Name()
					o[i].Size = ns[i].Size()
					o[i].LastModifiedDateTime = ns[i].ModTime()

					// if it, ok := ns[i].(share.Item); ok {
					// 	em := it.GetItem()

					// 	o[i].Folder = em.Folder
					// 	o[i].ID = em.ID
					// } else {

					o[i].Folder = &struct {
						ChildCount int `json:"childCount"`
					}{}
					// 注：前端将 ID 作为 key，需要保证唯一，否则会窜目录
					o[i].ID = path.Join(`vfs:`, qpath, o[i].Name)

					// }
				}
				ic.d = &share.DriveChildren{Value: o}
				/*d, _ := json.Marshal(AppFolder{share.DriveChildren{Value: o}})
				a.Cache.Set(ik, d, maxAge)
				ic.b = d*/
			}
		} else {
			// 信息
			var it share.DriveItem
			it.Name = f.Name()
			it.Size = f.Size()
			it.LastModifiedDateTime = f.ModTime()
			it.ID = `vfs:` + qpath
			// 链接
			if d, ok := f.(vfs.File); ok {
				u, err := d.Url(c)
				if err != nil {
					return nil, err
				}
				a.Cache.Set(lk, util.StringToBytes(u), maxAge)
				ic.l = u
				it.ContentDownloadURL = u
			}
			d, _ := json.Marshal(AppFile{&it})
			a.Cache.Set(ik, d, maxAge)
			ic.b = d
		}
		return ic, nil
	})
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if ic, ok := v.(ictx); ok {
		fmt.Fprintf(os.Stdout, "[OVI] node: name=%q, size=%v, modTime=%q, isDir=%v, type=%T\n", ic.n.Name(), ic.n.Size(), ic.n.ModTime().Format(time.DateTime), ic.n.IsDir(), ic.n)
		w.Header()[hkCC] = CC
		if raw {
			if ic.l != `` {
				http.Redirect(w, r, ic.l, http.StatusFound)
				return
			}
			Error(w, `No download url found.`, http.StatusNotFound)
			return
		}
		if ic.b != nil {
			JsonBlob(w, ic.b, http.StatusOK)
			return
		} else if ic.d != nil {
			Json(w, AppFolder{ic.d, ic.d.OdataNextLink}, http.StatusOK)
			return
		}
	}
	Error(w, `vfs: not support`, http.StatusNotImplemented)

	/*h := w.Header()

	h[`Cache-Control`] = CC

	f, err := a.FS.Open(c, qpath)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if f.IsDir() {
		// 目录不支持跳转链接
		if raw {
			Error(w, `No download url found.`, http.StatusNotFound)
			return
		}
		if d, ok := f.(*share.PersonalObject); ok {
			// next := q.Get(`next`)
			// sort := q.Get(`sort`)

			ns, err := d.ListItem(c, ``)
			if err != nil {
				Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			Json(w, AppFolder{share.DriveChildren{Value: ns}}, http.StatusOK)
			return
		} else {
			Error(w, `vfs: not a dir`, http.StatusNotImplemented)
			return
		}
	}

	if d, ok := f.(vfs.File); ok {
		if raw {
			k := util.ConcatB(`raw:`, qpath)
			if v, err := a.Cache.Get(k); err == nil {
				http.Redirect(w, r, util.BytesToString(v), http.StatusFound)
				return
			}
			u, err := d.Url(c)
			if err != nil {
				Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			a.Cache.Set(k, util.StringToBytes(u), 3600)
			http.Redirect(w, r, u, http.StatusFound)
			return
		}
		if o, ok := f.(*share.PersonalObject); ok {
			Json(w, AppFile{o.Item}, http.StatusOK)
			return
		}
		Error(w, `vfs: not support`, http.StatusNotImplemented)
		return
	} else {
		Error(w, `vfs: not a file`, http.StatusNotImplemented)
		return
	}*/
}
