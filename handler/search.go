package handler

import "net/http"

// 搜索（暂不支持）
func (a *Handler) Search(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
}
