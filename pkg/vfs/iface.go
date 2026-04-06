package vfs

import (
	"context"
	"fmt"
	"time"
)

type Node interface {
	Name() string
	Size() int64
	ModTime() time.Time
	IsDir() bool
}

func String(n Node) string {
	return fmt.Sprintf("[vfs.Node] name=%q, size=%v, modTime=%q, isDir=%v, type=%T", n.Name(), n.Size(), n.ModTime().Format(time.DateTime), n.IsDir(), n)
}

type File interface {
	Node

	// 获取文件外链
	Url(ctx context.Context) (string, error)
}

type Dir interface {
	Node

	// 列出文件
	List(ctx context.Context, subPath string) ([]Node, error)
}

type Thumb interface {
	Thumb(ctx context.Context, subPath string) ([]string, error)
}

// 文件系统提供者
type Provider interface {
	Node

	Open(ctx context.Context, subPath string) (Node, error)

	// ListDir(ctx context.Context, subPath string) ([]FileHandle, error)
	// GetFile(ctx context.Context, subPath string) (FileHandle, error)
}
