package util

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AI 生成
// https://grok.com/share/bGVnYWN5LWNvcHk_288dbb3d-3f46-47c3-b36a-e9125e228ba5

// TokenManager 是一个线程安全的 Token 管理器
// - 支持高并发获取 Token（有效 Token 时多个 goroutine 可同时读取，无阻塞）
// - 刷新 Token 时保证「单线程」执行（使用互斥锁 + 双重检查，防止惊群效应）
type TokenManager struct {
	token     string         // 当前 Token
	expireSec int64          // Token 过期时间
	Refresh   TkmRefreshFunc // 用户提供的刷新函数

	mu        sync.RWMutex // 保护 token 和 expiresAt，支持并发读
	refreshMu sync.Mutex   // 确保同一时刻只有一个 goroutine 在刷新
}

// TkmRefreshFunc 是用户需要实现的刷新逻辑
// 返回：新 Token、有效时长（TTL）、错误
// 建议在刷新函数里使用传入的 context 做 HTTP/GRPC 调用，支持超时/取消
type TkmRefreshFunc func(ctx context.Context) (newToken string, expire int64, err error)

// NewTokenManager 创建 Token 管理器
// refreshFunc 必须非 nil
func NewTokenManager(refreshFunc TkmRefreshFunc) *TokenManager {
	if refreshFunc == nil {
		panic("refreshFunc cannot be nil")
	}
	return &TokenManager{
		Refresh: refreshFunc,
	}
}

// GetToken 并发安全地获取 Token
// 如果 Token 有效，直接返回；如果过期，只有第一个到达的 goroutine 会执行刷新，其他等待
func (tm *TokenManager) GetToken(ctx context.Context) (string, error) {
	// 1. 快速路径：读锁检查是否有效（高并发友好）
	tm.mu.RLock()
	if tm.expireSec != 0 && time.Now().Unix() < tm.expireSec {
		token := tm.token
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	// 2. Token 已过期或首次使用，需要刷新
	// 使用 refreshMu 保证同一时刻只有一个 goroutine 执行刷新
	tm.refreshMu.Lock()
	defer tm.refreshMu.Unlock()

	// 3. 双重检查：可能在等待锁期间已经被其他 goroutine 刷新完成
	tm.mu.RLock()
	if tm.expireSec != 0 && time.Now().Unix() < tm.expireSec {
		token := tm.token
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	// 4. 执行真正的刷新（单线程）
	newToken, expire, err := tm.Refresh(ctx)
	if err != nil {
		return "", fmt.Errorf("refresh token failed: %w", err)
	}

	// 5. 更新 Token（写锁）
	tm.mu.Lock()
	tm.token = newToken
	tm.expireSec = expire
	tm.mu.Unlock()

	return newToken, nil
}

// 设置内部令牌
func (tm *TokenManager) SetToken(newToken string, expire int64) {
	tm.mu.Lock()
	tm.token = newToken
	tm.expireSec = expire
	tm.mu.Unlock()
}

// 获取内部令牌
func (tm *TokenManager) OldToken() (token string, expire int64) {
	tm.mu.RLock()
	/*if tm.expireSec != 0 && time.Now().Unix() < tm.expireSec {
		token = tm.token
	}*/
	token = tm.token
	expire = tm.expireSec
	tm.mu.RUnlock()
	return
}
