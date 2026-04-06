package share

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ZxwyWebSite/ovi-share/pkg/util"
)

// 解析 JWT 令牌过期时间
func parseExpiFromToken(tk string) int64 {
	tk = tk[3:]
	i := strings.IndexByte(tk, '.')
	if i != -1 {
		dec, err := util.Base64Decode(base64.RawStdEncoding, util.StringToBytes(tk[:i]))
		if err == nil {
			var meta tokenMeta
			err = json.Unmarshal(dec, &meta)
			if err == nil {
				exp, err := strconv.Atoi(meta.Exp)
				if err == nil {
					return int64(exp)
				}
			}
		}
	}
	return 0
}

type tokenMeta struct {
	// Siteid string `json:"siteid"`
	// Aud string `json:"aud"`

	Exp string `json:"exp"`
}
