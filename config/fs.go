package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ZxwyWebSite/ovi-share/pkg/util"
	"github.com/ZxwyWebSite/ovi-share/pkg/vfs"
	"github.com/ZxwyWebSite/ovi-share/provider/share"
)

type getMetaFunc func(name, prefix string, isLast bool) (vfs.Provider, error)

func (c *Config) Build(ctx context.Context) (vfs.Provider, map[string]vfs.Provider, error) {
	ctx = context.WithValue(ctx, `conf`, c)
	// Meta 懒加载
	l := len(c.Meta)
	m := make(map[string]vfs.Provider)
	var getMeta getMetaFunc
	getMeta = func(name, prefex string, isLast bool) (vfs.Provider, error) {
		if p, ok := m[name]; ok {
			return p, nil
		}
		for i := 0; i < l; i++ {
			if c.Meta[i].Name == name {
				p, err := buildFS(ctx, &c.Meta[i], getMeta, prefex, isLast, true)
				if err != nil {
					return nil, fmt.Errorf(`meta[%d:%s]: %w`, i, c.Meta[i].Name, err)
				}
				m[c.Meta[i].Name] = p
				return p, nil
			}
		}
		return nil, fmt.Errorf(`undefined ref: %s`, name)
	}
	// 构建根目录
	os.Stdout.WriteString("root\n")
	p, err := buildFS(ctx, c.Root, getMeta, ``, true, false)
	if err != nil {
		return nil, nil, fmt.Errorf(`vfs.Build: %w`, err)
	}
	// 构建站点目录
	k := len(c.Site)
	s := make(map[string]vfs.Provider)
	for i := 0; i < k; i++ {
		os.Stdout.Write(util.ConcatB("\n", c.Site[i].Name, "\n"))
		p, err := buildFS(ctx, &c.Site[i], getMeta, ``, true, false)
		if err != nil {
			return nil, nil, fmt.Errorf(`vfs.Build: site[%d:%s]: %w`, i, c.Site[i].Name, err)
		}
		s[c.Site[i].Name] = p
	}
	return p, s, nil
}

// 递归构建
func buildFS(ctx context.Context, p *Provider, getMeta getMetaFunc,
	prefix string, isLast, isMeta bool,
) (vfs.Provider, error) {
	// 确定连接符
	var connector string
	if isMeta {
		connector = "├─# "
	} else if isLast {
		connector = "└── "
	} else {
		connector = "├── "
	}

	// 构建当前项
	var o vfs.Provider
	switch p.Type {
	case `share`:
		// OneDrive 分享
		if p.Share != nil {
			var s *share.Object
			var err error
			conf := ctx.Value(`conf`).(*Config)
			if strings.Contains(p.Share.Link, `sharepoint`) {
				// 商业版
				c := share.NewBusiness()

				var bf *share.BusinessFetch
				if p.Share.Token == `` || time.Now().After(time.Unix(p.Share.Expire, 0)) || p.Share.Root == `` || p.Share.Path == `` {
					bf, err = c.Fetch(ctx, p.Share.Link)
				} else {
					bf = &share.BusinessFetch{
						Token:  p.Share.Token,
						Expire: p.Share.Expire,
						Root:   p.Share.Root,
						Path:   p.Share.Path,
					}
				}
				if err == nil {
					fn := c.TM.Refresh
					c.TM.Refresh = func(ctx context.Context) (newToken string, expire int64, err error) {
						newToken, expire, err = fn(ctx)
						if err == nil {
							p.Share.Token = newToken
							p.Share.Expire = expire
							conf.Save(``)
						}
						return
					}
					s, err = c.ObjectRaw(ctx, p.Share.Link, bf)
					if err == nil && (p.Share.Token != bf.Token || p.Share.Expire != bf.Expire || p.Share.Root != bf.Root || p.Share.Path != bf.Path) {
						p.Share.Token = bf.Token
						p.Share.Expire = bf.Expire
						p.Share.Root = bf.Root
						p.Share.Path = bf.Path
						conf.Save(``)
					}
				}
			} else {
				// 个人版
				c := share.NewPersonal()

				c.TM.SetToken(p.Share.Token, p.Share.Expire)
				fn := c.TM.Refresh
				c.TM.Refresh = func(ctx context.Context) (newToken string, expire int64, err error) {
					newToken, expire, err = fn(ctx)
					if err == nil {
						p.Share.Token = newToken
						p.Share.Expire = expire
						conf.Save(``)
					}
					return
				}
				s, err = c.Object(ctx, p.Share.Link)
			}
			if err != nil {
				return nil, err
			}
			o = s
		}
	case `mount`:
		// 打印当前行
		os.Stdout.Write(util.ConcatB(prefix, connector, p.Name, `: `, p.Type, "\n"))

		// 计算下一层的前缀
		var newPrefix string
		if isMeta {
			newPrefix = prefix + "│ # "
		} else if isLast {
			newPrefix = prefix + "    "
		} else {
			newPrefix = prefix + "│   "
		}

		// 虚拟目录挂载（允许空目录？）
		c := vfs.NewMountFS(p.Name)
		l := len(p.Mount)
		for i := 0; i < l; i++ {
			isLastChild := i == l-1
			f, err := buildFS(ctx, &p.Mount[i], getMeta, newPrefix, isLastChild, false)
			if err != nil {
				return nil, fmt.Errorf(`mount[%d:%s]: %w`, i, p.Mount[i].Name, err)
			}
			c.Mount(p.Mount[i].Name, f)
		}
		c.Build()
		return c, nil
	case `ref`:
		// 元数据引用
		r, err := getMeta(p.Ref, prefix, isLast)
		if err != nil {
			return nil, err
		}
		o = r
	default:
		return nil, fmt.Errorf(`unknown type %q`, p.Type)
	}
	if o == nil {
		return nil, errors.New(`invalid configuration`)
	}

	// 打印当前行
	os.Stdout.Write(util.ConcatB(prefix, connector, p.Name, `: `, p.Type, ` - `, vfs.String(o), "\n"))

	return o, nil
}
