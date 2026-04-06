package handler

import (
	"context"
	"net/http"
)

// 跳转链接（别名）
func (a *Handler) Raw(w http.ResponseWriter, r *http.Request) {
	a.Index(w, r.WithContext(context.WithValue(r.Context(), `raw`, true)))
}
