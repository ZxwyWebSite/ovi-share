package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// From ECHO

type CorsConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           int
}

var allMethods = [...]string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete}

func (c *CorsConfig) ToMiddleware() func(http.Handler) http.HandlerFunc {
	allowOriginLen := len(c.AllowOrigins)
	allowOriginMap := make(map[string]struct{}, allowOriginLen)
	allowOriginAll := false
	for i := 0; i < allowOriginLen; i++ {
		if c.AllowOrigins[i] == `*` {
			allowOriginAll = true
			allowOriginMap = nil
			break
		}
		allowOriginMap[c.AllowOrigins[i]] = struct{}{}
	}

	if len(c.AllowMethods) == 0 {
		c.AllowMethods = allMethods[:]
	}

	var allowMethods, allowHeaders, maxAge, allowCredentials []string

	allowMethods = []string{strings.Join(c.AllowMethods, `,`)}

	if len(c.AllowHeaders) != 0 {
		allowHeaders = []string{strings.Join(c.AllowHeaders, `,`)}
	}

	if c.MaxAge > 0 {
		maxAge = []string{strconv.Itoa(c.MaxAge)}
	}
	if c.AllowCredentials {
		allowCredentials = []string{`true`}
	}

	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()

			header[`Vary`] = append(header[`Vary`], `Origin`)

			preflight := r.Method == http.MethodOptions

			origin := r.Header[`Origin`]
			if len(origin) == 0 || origin[0] == "" {
				if preflight {
					w.WriteHeader(http.StatusNoContent)
				} else {
					next.ServeHTTP(w, r)
				}
				return
			}

			if allowOriginAll {
				header[`Access-Control-Allow-Origin`] = origin
			} else if _, ok := allowOriginMap[origin[0]]; ok {
				header[`Access-Control-Allow-Origin`] = origin
			} else {
				if preflight {
					w.WriteHeader(http.StatusNoContent)
				} else {
					next.ServeHTTP(w, r)
				}
				return
			}

			if allowCredentials != nil {
				header[`Access-Control-Allow-Credentials`] = allowCredentials
			}

			if !preflight {
				next.ServeHTTP(w, r)
				return
			}

			header[`Vary`] = append(header[`Vary`],
				`Access-Control-Request-Method`,
				`Access-Control-Request-Headers`,
			)

			header[`Access-Control-Allow-Methods`] = allowMethods

			if allowHeaders != nil {
				header[`Access-Control-Allow-Headers`] = allowHeaders
			} else {
				h := r.Header[`Access-Control-Request-Headers`]
				if len(h) != 0 && h[0] != "" {
					header[`Access-Control-Allow-Headers`] = h
				}
			}
			if maxAge != nil {
				header[`Access-Control-Max-Age`] = maxAge
			}

			w.WriteHeader(http.StatusNoContent)
		}
	}
}
