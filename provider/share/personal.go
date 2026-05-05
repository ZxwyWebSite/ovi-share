package share

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/ZxwyWebSite/ovi-share/pkg/util"
	"github.com/ZxwyWebSite/ovi-share/pkg/vfs"
	"golang.org/x/sync/errgroup"
)

// 个人版
type Personal struct {
	Client *http.Client
	TM     util.TokenManager
	IDS    []string // 记录资源链接，刷新令牌后恢复绑定
}

// 创建个人版实例（注意有交叉引用，必须在此处创建）
func NewPersonal() *Personal {
	c := new(Personal)
	c.Client = http.DefaultClient
	c.TM.Refresh = c.getToken
	return c
}

/// TokenRefresh

type personalToken struct {
	AuthScheme    string    `json:"authScheme"` // badger
	Token         string    `json:"token"`
	ExpiryTimeUtc time.Time `json:"expiryTimeUtc"`
}

// 刷新访问令牌
func (c *Personal) getToken(ctx context.Context) (string, int64, error) {
	const body = `{"appId":"00000000-0000-0000-0000-0000481710a4"}`

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, `https://api-badgerp.svc.ms/v1.0/token`, strings.NewReader(body))
	req.Header[ctKey] = ctJsonOdata
	req.Header[uaKey] = uaFireFox

	res, err := c.Client.Do(req)
	if err == nil {
		if res.StatusCode != http.StatusOK {
			err = fmt.Errorf(`share.Personal.getToken: unexpected status: %s`, res.Status)
			res.Body.Close()
		} else {
			var out personalToken
			err = json.NewDecoder(res.Body).Decode(&out)
			res.Body.Close()
			if err == nil {
				expire := out.ExpiryTimeUtc.Unix()
				if expire != 0 {
					expire -= 60
				}
				token := out.AuthScheme + ` ` + out.Token
				// rebind
				if l := len(c.IDS); l != 0 {
					eg, ec := errgroup.WithContext(ctx)
					for i := 0; i < l; i++ {
						i := i
						eg.Go(func() error {
							_, err := c.bind(ec, token, c.IDS[i])
							return err
						})
					}
					err = eg.Wait()
				}
				return token, expire, err
			}
		}
	}
	return ``, 0, err
}

// 获取访问令牌
func (c *Personal) GetToken() (string, error) {
	return c.TM.GetToken(context.Background())
}

/// Init

const personalRoot = `https://my.microsoftpersonalcontent.com/_api/v2.0`

// 内部绑定实现
func (c *Personal) bind(ctx context.Context, token, link string) (*DriveItem, error) {
	uri := util.Concat(personalRoot, `/shares/u!`, util.BytesToString(util.Base64Encode(base64.RawURLEncoding, util.StringToBytes(link))), `/driveitem`)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	req.Header[`Authorization`] = []string{token}
	req.Header[uaKey] = uaFireFox
	req.Header[`Prefer`] = prefer // ※ 不加这个无法将分享绑定到令牌

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	var out DriveItem
	err = json.NewDecoder(res.Body).Decode(&out)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// 初始化分享信息
func (c *Personal) Object(ctx context.Context, link string) (*Object, error) {
	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}

	item, err := c.bind(ctx, token, link)
	if err != nil {
		return nil, err
	}
	c.IDS = append(c.IDS, link)

	// 打开分享的文件或目录
	return (&Object{
		Client:   c.Client,
		GetToken: c.GetToken,
		Root:     personalRoot,

		Item: item,
	}).OpenRaw(ctx, ``)
}

/// Object

// 可复用实现
type Object struct {
	Client   *http.Client
	GetToken func() (string, error)
	Root     string // API 端点

	Item *DriveItem
}

func (c *Object) Name() string       { return c.Item.Name }
func (c *Object) Size() int64        { return c.Item.Size }
func (c *Object) ModTime() time.Time { return c.Item.LastModifiedDateTime }
func (c *Object) IsDir() bool        { return c.Item.ContentDownloadURL == `` }
func (c *Object) Sys() any           { return c.Item }

func (c *Object) GetItem() *DriveItem { return c.Item }

var _ vfs.Node = (*Object)(nil)
var _ vfs.Dir = (*Object)(nil)
var _ vfs.File = (*Object)(nil)
var _ Item = (*Object)(nil)

// 合并子目录
func (c *Object) SubPath(sub, end string) string {
	// 不编码前缀
	// personal: /drives/XXXXXXXXXXXXXXXX/root:
	// business: /drives/XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX/root:
	off := 14 + len(c.Item.ParentReference.DriveID)
	return util.Concat(c.Root, c.Item.ParentReference.Path[:off], url.PathEscape(path.Join(c.Item.ParentReference.Path[off:], `/`, c.Item.Name, sub)), `:`, end)
}

// 列出相对当前目录
func (c *Object) ListItem(ctx context.Context, subPath, next string) (*DriveChildren, error) {
	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}

	var uri string
	if subPath == `` || subPath == `/` || subPath == `.` {
		if next != `` {
			uri = util.Concat(c.Root, `/drives/`, c.Item.ParentReference.DriveID, `/items/`, c.Item.ID, `/children`, `?$skipToken=`, next)
		} else {
			uri = util.Concat(c.Root, `/drives/`, c.Item.ParentReference.DriveID, `/items/`, c.Item.ID, `/children`)
		}
	} else {
		// uri = util.Concat(c.Root, c.Item.ParentReference.Path[:30], c.SubPath(subPath), `:/children`)
		if next != `` {
			off := 14 + len(c.Item.ParentReference.DriveID)
			uri = util.Concat(c.Root, c.Item.ParentReference.Path[:off], url.PathEscape(path.Join(c.Item.ParentReference.Path[off:], `/`, c.Item.Name, subPath)), `:`, `/children`, `?$skipToken=`, next)
		} else {
			uri = c.SubPath(subPath, `/children`)
		}
	}
	// println(`list:`, uri)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	req.Header[`Authorization`] = []string{token}
	req.Header[uaKey] = uaFireFox

	// dump, _ := httputil.DumpRequest(req, false)
	// os.Stdout.Write(dump)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// dump, _ = httputil.DumpResponse(res, false)
	// os.Stdout.Write(dump)

	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf(`share.Object.List: unexpected status: %s`, res.Status)
	}

	var out DriveChildren
	err = json.NewDecoder(res.Body).Decode(&out)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// 列出相对当前目录
func (c *Object) ListRaw(ctx context.Context, subPath string) ([]*Object, error) {
	items, err := c.ListItem(ctx, subPath, ``)
	if err == nil {
		if l := len(items.Value); l != 0 {
			obj := make([]*Object, l)
			for i := 0; i < l; i++ {
				obj[i] = &Object{
					Client:   c.Client,
					GetToken: c.GetToken,
					Root:     c.Root,
					Item:     &items.Value[i],
				}
			}
			return obj, nil
		}
	}
	return nil, err
}

// 列出相对当前目录（VFS）
func (c *Object) List(ctx context.Context, subPath string) ([]vfs.Node, error) {
	objs, err := c.ListRaw(ctx, subPath)
	if err == nil {
		if l := len(objs); l != 0 {
			o := make([]vfs.Node, l)
			for i := 0; i < l; i++ {
				o[i] = objs[i]
			}
			return o, nil
		}
	}
	return nil, err
}

// 打开相对当前目录
func (c *Object) OpenRaw(ctx context.Context, subPath string) (*Object, error) {
	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}

	var uri string
	if c.Item.ID != `` && (subPath == `` || subPath == `/` || subPath == `.`) {
		uri = util.Concat(c.Root, `/drives/`, c.Item.ParentReference.DriveID, `/items/`, c.Item.ID)
	} else {
		uri = c.SubPath(subPath, ``)
	}
	/*switch subPath {
	case ``, `/`, `.`:
		uri = util.Concat(c.Root, `/drives/`, c.Item.ParentReference.DriveID, `/items/`, c.Item.ID)
	default:
		uri = util.Concat(c.Root, c.Item.ParentReference.Path[:30], c.SubPath(subPath))
	}*/
	// println(`open:`, uri)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	req.Header[`Authorization`] = []string{token}
	req.Header[uaKey] = uaFireFox

	// dump, _ := httputil.DumpRequest(req, false)
	// os.Stdout.Write(dump)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// dump, _ = httputil.DumpResponse(res, false)
	// os.Stdout.Write(dump)

	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf(`share.Object.Open: unexpected status: %s`, res.Status)
	}

	var out DriveItem
	err = json.NewDecoder(res.Body).Decode(&out)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	return &Object{
		Client:   c.Client,
		GetToken: c.GetToken,
		Root:     c.Root,
		Item:     &out,
	}, nil
}

// 打开相对当前目录
func (c *Object) Open(ctx context.Context, subPath string) (vfs.Node, error) {
	return c.OpenRaw(ctx, subPath)
}

// 获取当前对象链接（VFS）
func (c *Object) Url(ctx context.Context) (string, error) {
	return c.Item.ContentDownloadURL, nil
}

// 获取当前对象链接（TODO：缓存链接过期处理）
func (c *Object) UrlNew(ctx context.Context) (string, error) {
	o, err := c.OpenRaw(ctx, ``)
	if err != nil {
		return ``, err
	}
	return o.Item.ContentDownloadURL, nil
}

// 获取缩略图（VFS）
func (c *Object) Thumb(ctx context.Context, subPath string) ([]string, error) {
	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}

	var uri string
	if subPath == `` || subPath == `/` || subPath == `.` {
		uri = util.Concat(c.Root, `/drives/`, c.Item.ParentReference.DriveID, `/items/`, c.Item.ID, `/thumbnails`)
	} else {
		// uri = util.Concat(c.Root, c.Item.ParentReference.Path[:30], c.SubPath(subPath), `:/thumbnails`)
		uri = c.SubPath(subPath, `/thumbnails`)
	}
	// println(`thumb:`, uri)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	req.Header[`Authorization`] = []string{token}
	req.Header[uaKey] = uaFireFox

	// dump, _ := httputil.DumpRequest(req, false)
	// os.Stdout.Write(dump)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// dump, _ := httputil.DumpResponse(res, false)
	// os.Stdout.Write(dump)

	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf(`share.Object.Thumb: unexpected status: %s`, res.Status)
	}

	var out DriveThumbs
	err = json.NewDecoder(res.Body).Decode(&out)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	if len(out.Value) != 0 {
		o := make([]string, 3)
		o[0] = out.Value[0].Small.URL
		o[1] = out.Value[0].Medium.URL
		o[2] = out.Value[0].Large.URL
		return o, nil
	}

	return nil, nil
}
