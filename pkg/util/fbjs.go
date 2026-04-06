package util

import (
	"io/fs"
	"os"
	"path/filepath"
)

// 默认权限
/*
  0 | 所有者 | 用户组 | 公共

 读取 (4): f t t t f f f t
 写入 (2): f f t t t f t f
 执行 (1): f f f t t t f t
 权限 (+): 0 4 6 7 3 1 2 5

 R: × × × × √ √ √ √
 W: × × √ √ × × √ √
 X: × √ × √ × √ × √
 P: 0 1 2 3 4 5 6 7
*/
var DefPerm fs.FileMode = 0666 //0644
var DirPerm fs.FileMode = 0755

// 判断文件是否存在 <文件路径(相对或绝对皆可)> <存在?>
func IsExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// 为文件创建上级目录
func MkdirAll(path string, perm fs.FileMode) error {
	if dir := filepath.Dir(path); !IsExists(dir) {
		return os.MkdirAll(dir, perm)
	}
	return nil
}
func Mkdir(path string) error {
	return MkdirAll(path, DirPerm)
}

// 打开文件 <路径, 标识, 权限> <*文件, 错误>
/*
 用法同 os.OpenFile
 调用前会检测文件目录是否存在，否则补充创建
*/
func OpenFile(path string, flag int, perm fs.FileMode) (*os.File, error) {
	if err := Mkdir(path); err != nil {
		return nil, err
	}
	return os.OpenFile(path, flag, perm)
}

// 快速创建文件 (同 os.Create)
func CreatFile(path string) (*os.File, error) {
	return OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, DefPerm)
}

// 写入文件内容 (同 os.WriteFile)
func WriteFile(name string, data []byte, perm fs.FileMode) error {
	f, err := OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}
func SaveFile(name string, data []byte) error {
	return WriteFile(name, data, DefPerm)
}
