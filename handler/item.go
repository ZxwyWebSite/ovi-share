package handler

import "net/http"

// 指定对象（多目录挂载时无法支持）
func (a *Handler) Item(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
}
