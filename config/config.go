package config

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/ZxwyWebSite/ovi-share/pkg/util"
)

// 主配置
type Config struct {
	// 载入路径
	path string
	mu   sync.Mutex

	// 服务器
	Serv CfgServ `json:"serv"`
	// 元数据
	Meta []Provider `json:"meta"`
	// 根目录
	Root *Provider `json:"root"`
	// 站点
	Site []Provider `json:"site"`
}

type CfgServ struct {
	// 监听地址
	Listen string `json:"listen"`
	// 缓存大小（MB）
	Cache int `json:"cache"`
	// 静态文件
	Static string `json:"static"`
	// 跨站配置
	Cors CfgCors `json:"cors"`
}

type CfgCors struct {
	Enable       bool     `json:"enable"`
	AllowOrigins []string `json:"allowOrigins"`
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

	Token  string `json:"token,omitempty"`
	Expire int64  `json:"expire,omitempty"`

	Root string `json:"root,omitempty"`
	Path string `json:"path,omitempty"`
}

var Default = Config{
	Serv: CfgServ{
		Listen: `:1122`,
		Cache:  16,
		Static: `data/build`,
		Cors: CfgCors{
			AllowOrigins: []string{`*`},
		},
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
		mi, _ := json.MarshalIndent(&Default, ``, `  `)
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

	config.path = cfg
	return &config, nil
}

// 保存
func (c *Config) Save(to string) error {
	if to == `` {
		to = c.path
	}
	mi, _ := json.MarshalIndent(c, ``, `  `)
	c.mu.Lock()
	err := util.SaveFile(to, mi)
	c.mu.Unlock()
	return err
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
	if cfg.Serv.Cors.AllowOrigins == nil {
		cfg.Serv.Cors.AllowOrigins = Default.Serv.Cors.AllowOrigins
		err = ErrSave
	}
	if cfg.Root == nil {
		err = ErrRoot
	}
	return
}
