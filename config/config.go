package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ZxwyWebSite/ovi-share/pkg/util"
	"github.com/ZxwyWebSite/ovi-share/pkg/vfs"
	"github.com/ZxwyWebSite/ovi-share/provider/share"
)

// 主配置
type Config struct {
	// 服务器
	Serv CfgServ `json:"serv"`
	// 元数据
	Meta []Provider `json:"meta"`
	// 根目录
	Root *Provider `json:"root"`
}

type CfgServ struct {
	// 监听地址
	Listen string `json:"listen"`
	// 缓存大小（MB）
	Cache int `json:"cache"`
	// 静态文件
	Static string `json:"static"`
}

type Provider struct {
	Type string `json:"type"`
	Name string `json:"name"`

	// 分享
	Share *CfgShare `json:"share,omitempty"`
	// 引用
	Ref string `json:"ref,omitempty"`
	// 挂载目录
	Mount []Provider `json:"mount,omitempty"`
}

type CfgShare struct {
	Link string `json:"link"`
}

var Default = Config{
	Serv: CfgServ{
		Listen: `:1122`,
		Cache:  16,
		Static: `data/build`,
	},
	Root: &Provider{
		Type:  `mount`,
		Name:  `/`,
		Mount: []Provider{},
	},
}

var ErrInit = errors.New(`配置初始化成功，请编辑后重新运行`)

// 载入
func Load(cfg string) (*Config, error) {
	if !util.IsExists(cfg) {
		mi, _ := json.MarshalIndent(Default, ``, `  `)
		err := util.SaveFile(cfg, mi)
		if err == nil {
			err = ErrInit
		}
		return nil, err
	}

	data, err := os.ReadFile(cfg)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	err = Valid(&config)
	if err != nil {
		if err == ErrSave {
			mi, _ := json.MarshalIndent(&config, ``, `  `)
			err = util.SaveFile(cfg, mi)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &config, nil
}

var ErrSave = errors.New(`save`)

var ErrRoot = errors.New(`未配置根目录`)

func Valid(cfg *Config) (err error) {
	if cfg.Serv.Listen == `` {
		cfg.Serv.Listen = Default.Serv.Listen
		err = ErrSave
	}
	if cfg.Serv.Cache <= 0 {
		cfg.Serv.Cache = Default.Serv.Cache
		err = ErrSave
	}
	if cfg.Serv.Static == `` {
		cfg.Serv.Static = Default.Serv.Static
		err = ErrSave
	}
	if cfg.Root == nil {
		err = ErrRoot
	}
	return
}

func (c *Config) Build(ctx context.Context) (vfs.Provider, error) {
	// 首先构建引用
	println(`// Meta:`)
	l := len(c.Meta)
	m := make(map[string]vfs.Provider)
	for i := 0; i < l; i++ {
		p, err := buildFS(ctx, &c.Meta[i], m)
		if err != nil {
			return nil, fmt.Errorf(`vfs.Build: meta[%d]: %w`, i, err)
		}
		m[c.Meta[i].Name] = p
	}
	// 然后构建目录
	println(`// Root:`)
	p, err := buildFS(ctx, c.Root, m)
	if err != nil {
		return nil, fmt.Errorf(`vfs.Build: %w`, err)
	}
	return p, nil
}

// 递归构建
func buildFS(ctx context.Context, p *Provider, m map[string]vfs.Provider) (vfs.Provider, error) {
	switch p.Type {
	case `share`:
		// OneDrive 分享
		if p.Share != nil {
			var s *share.Object
			var err error
			if strings.Contains(p.Share.Link, `sharepoint`) {
				// 商业版
				c := share.NewBusiness()
				s, err = c.Object(ctx, p.Share.Link)
			} else {
				// 个人版
				c := share.NewPersonal()
				s, err = c.Object(ctx, p.Share.Link)
			}
			if err != nil {
				return nil, err
			}
			println(vfs.String(s))
			return s, nil
		}
	case `mount`:
		// 虚拟目录挂载（允许空目录？）
		// if p.Mount != nil {
		c := vfs.NewMountFS()
		l := len(p.Mount)
		for i := 0; i < l; i++ {
			f, err := buildFS(ctx, &p.Mount[i], m)
			if err != nil {
				return nil, fmt.Errorf(`mount[%d]: %w`, i, err)
			}
			c.Mount(p.Mount[i].Name, f)
			println(vfs.String(f))
		}
		c.Build()
		println(vfs.String(c))
		return c, nil
		// }
	case `ref`:
		// 元数据引用
		if r, ok := m[p.Ref]; ok {
			println(vfs.String(r))
			return r, nil
		}
		return nil, fmt.Errorf(`undefined ref: %s`, p.Ref)
	default:
		return nil, fmt.Errorf(`unknown type %q`, p.Type)
	}
	return nil, errors.New(`invalid configuration`)
}
