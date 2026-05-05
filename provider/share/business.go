package share

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ZxwyWebSite/ovi-share/pkg/util"
)

// 企业版
type Business struct {
	Client *http.Client
	TM     util.TokenManager
	Link   string
}

// 创建企业版实例
func NewBusiness() *Business {
	c := new(Business)
	client := new(http.Client)
	client.CheckRedirect = businessCheckRedirect
	c.Client = client
	c.TM.Refresh = c.getToken
	return c
}

func businessCheckRedirect(req *http.Request, via []*http.Request) error {
	// 简单传递 Cookie
	if req.Response != nil {
		cookies := req.Response.Cookies()
		if len(cookies) != 0 {
			cookie := cookies[0]
			req.Header[`Cookie`] = []string{cookie.Name + `=` + cookie.Value}
		}
	}
	return nil
}

/// Token

func (c *Business) getToken(ctx context.Context) (string, int64, error) {
	bf, err := c.Fetch(ctx, c.Link)
	if err == nil {
		return bf.Token, bf.Expire, nil
	}
	return ``, 0, nil
}

// 获取访问令牌
func (c *Business) GetToken() (string, error) {
	return c.TM.GetToken(context.Background())
}

/// Init

type BusinessFetch struct {
	Token  string
	Expire int64

	Root string
	Path string
}

// 从页面获取分享信息
func (c *Business) Fetch(ctx context.Context, link string) (*BusinessFetch, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	req.Header[uaKey] = uaFireFox

	var bf BusinessFetch

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	var dInfo shareDirInfo

	err = indexStreamBufferSpecial(res.Body, util.MakeNoZero(4096), func(i int, buf []byte) (bool, error) {
		d := json.NewDecoder(io.MultiReader(bytes.NewReader(buf), res.Body))
		switch i {
		case 1:
			err := d.Decode(&dInfo)
			if err != nil || dInfo.DriveInfo.DriveAccessToken == `` {
				return false, err
			}
			id := res.Request.URL.Query().Get(`id`)
			if id == `` {
				return false, errors.New(`failed to get share root`)
			}
			// 取分享根路径
			idx := strings.Index(id, `/Documents`)
			if idx == -1 {
				return false, errors.New(`failed to index share root`)
			}
			id = id[idx+10:]

			bf.Token = `Bearer ` + dInfo.DriveInfo.DriveAccessToken[13:] // access_token=v1...
			// 注：虽然没有过期，但一段时间不活动会失效
			bf.Expire = time.Now().Unix() + 3500 //parseExpiFromToken(dInfo.DriveInfo.DriveAccessToken[13:])
			// 拼接 API 地址
			bf.Root = util.Concat(res.Request.URL.Scheme, `://`, res.Request.URL.Host, `/_api/v2.0`)
			// 拼接父目录地址
			bf.Path = util.Concat(dInfo.DriveInfo.DriveURL[len(bf.Root):], `/root:`, id)
		}
		return true, nil
	})
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	// fmt.Printf("bf: %#v\n", bf)

	return &bf, nil
}

/// Object

// 从页面提取令牌，获取分享目录信息（注意企业版只能单例）
func (c *Business) Object(ctx context.Context, link string) (*Object, error) {
	bf, err := c.Fetch(ctx, link)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("bf: %#v\n", bf)
	return c.ObjectRaw(ctx, link, bf)
}

// 从已知数据初始化
func (c *Business) ObjectRaw(ctx context.Context, link string, bf *BusinessFetch) (*Object, error) {
	c.TM.SetToken(bf.Token, bf.Expire)
	c.Link = link

	var it DriveItem
	// 占位符，仅用于长度标识，不另做提取
	it.ParentReference.DriveID = `XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX`
	it.ParentReference.Path = bf.Path

	return (&Object{
		Client:   c.Client,
		GetToken: c.GetToken,
		Root:     bf.Root,
		Item:     &it,
	}).OpenRaw(ctx, ``)
}
