package vfs

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"
)

type MountFS struct {
	name    string
	modTime time.Time

	// 挂载点: /drive, /s3, /local
	// mounts map[string]Provider

	// 根目录缓存（排序后的节点列表）
	nodes []MountPoint
}

func NewMountFS() *MountFS {
	return &MountFS{
		name:    "/",
		modTime: time.Now(),
		// mounts:  make(map[string]Provider),
	}
}

// 挂载 Provider
func (m *MountFS) Mount(path string, p Provider) {
	path = strings.Trim(path, "/")
	// m.mounts[path] = p
	m.nodes = append(m.nodes, MountPoint{
		name:     path,
		Provider: p,
	})
}

func (m *MountFS) Name() string       { return m.name }
func (m *MountFS) Size() int64        { return 0 }
func (m *MountFS) ModTime() time.Time { return m.modTime }
func (m *MountFS) IsDir() bool        { return true }

func (m *MountFS) Open(ctx context.Context, subPath string) (Node, error) {
	subPath = path.Clean(subPath)
	subPath = strings.TrimPrefix(subPath, "/")

	// 根目录：列出所有挂载点
	if subPath == "" || subPath == "." {
		return &MountDir{
			name: "/",
			// mounts:  m.mounts,
			modTime: m.modTime,
			nodes:   m.nodes,
		}, nil
	}

	parts := strings.SplitN(subPath, "/", 2)
	mountName := parts[0]

	i, ok := sort.Find(len(m.nodes), func(i int) int {
		return strings.Compare(mountName, m.nodes[i].name)
	})
	// p, ok := m.mounts[mountName]
	if !ok {
		return nil, fmt.Errorf("mount not found")
	}
	p := m.nodes[i]

	// 子路径
	if len(parts) == 1 {
		return p.Provider, nil
	}

	return p.Open(ctx, parts[1])
}

// 排序
func (m *MountFS) Build() {
	sort.Slice(m.nodes, func(i, j int) bool {
		return m.nodes[i].name < m.nodes[j].name
	})
}

var ErrNotFound = errors.New(`vfs: mount not found`)
var ErrNotSupport = errors.New(`vfs: method not support`)

// 缩略图
func (m *MountFS) Thumb(ctx context.Context, subPath string) ([]string, error) {
	subPath = strings.TrimPrefix(subPath, "/")

	// 根目录：不支持
	if subPath == "" || subPath == "." {
		return nil, ErrNotSupport
	}

	parts := strings.SplitN(subPath, "/", 2)
	mountName := parts[0]

	i, ok := sort.Find(len(m.nodes), func(i int) int {
		return strings.Compare(mountName, m.nodes[i].name)
	})
	if !ok {
		return nil, ErrNotFound
	}
	p := m.nodes[i]

	// 子路径
	if len(parts) == 1 {
		return nil, ErrNotSupport
	}

	if t, ok := p.Provider.(Thumb); ok {
		return t.Thumb(ctx, parts[1])
	}

	return nil, ErrNotSupport
}

type MountDir struct {
	name    string
	modTime time.Time
	// mounts  map[string]Provider
	nodes []MountPoint
}

func (d *MountDir) Name() string       { return d.name }
func (d *MountDir) Size() int64        { return 0 }
func (d *MountDir) ModTime() time.Time { return d.modTime }
func (d *MountDir) IsDir() bool        { return true }

func (d *MountDir) List(ctx context.Context, subPath string) ([]Node, error) {
	l := len(d.nodes)
	nodes := make([]Node, l)
	for i := 0; i < l; i++ {
		nodes[i] = &d.nodes[i]
	}
	/*for name, p := range d.mounts {
		nodes = append(nodes, &MountPoint{
			name:     name,
			Provider: p,
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name() < nodes[j].Name()
	})*/
	return nodes, nil
}

type MountPoint struct {
	name string
	Provider
}

func (m *MountPoint) Name() string       { return m.name }
func (m *MountPoint) Size() int64        { return m.Provider.Size() }
func (m *MountPoint) ModTime() time.Time { return m.Provider.ModTime() }
func (m *MountPoint) IsDir() bool        { return true }
