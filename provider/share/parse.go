package share

import (
	"bytes"
	"io"

	"github.com/ZxwyWebSite/ovi-share/pkg/util"
)

// g_fileInfo
type shareFileInfo struct {
	Name      string `json:"name"`      // 全名
	Extension string `json:"extension"` // 扩展名
	Size      int64  `json:"size"`      // 大小

	// PermMask string `json:"PermMask"`
	// Title any `json:"title"`

	DisplayName string `json:"displayName"` // 文件名

	// ProgID string `json:"ProgId"`
	// ContentCategory string `json:"ContentCategory"`

	VroomItemID string `json:"VroomItemId"`

	// ListTemplateType string `json:"ListTemplateType"`

	DownloadURL string `json:"downloadUrl"` // 下载链接，签名有效期一小时

	DownloadURLNoAuth string `json:"downloadUrlNoAuth"` // 没有签名的下载链接

	IsDownloadBlocked bool `json:"isDownloadBlocked"`

	// Description any `json:"Description"`
	// CreatedByUser string `json:"CreatedByUser"` // 创建者用户名
	// CreatedDateTime time.Time `json:"CreatedDateTime"` // 文件创建时间
	// ExpirationDate any `json:"ExpirationDate"`
	// CommentCount int `json:"CommentCount"`
	// Path string `json:"path"`
	// PrincipalCount string `json:"PrincipalCount"`
	// WebURL string `json:"webUrl"`

	SpItemURL string `json:".spItemUrl"` // 对象信息链接，需要拼接 &tempauth={DriveAccessToken}

	// TransformURL string `json:".transformUrl"` // 媒体转码，暂时用不到

	DriveAccessToken string `json:".driveAccessToken"` // 临时 Token，有效期6小时（具体读jwt meta段的exp）

	// DriveAccessTokenV21 string `json:".driveAccessTokenV21"`
	// DriveAccessCode string `json:".driveAccessCode"`
	// DriveAccessCodeV21 string `json:".driveAccessCodeV21"`
	// Etag string `json:".etag"`
	// Ctag string `json:".ctag"`
	// MediaServiceFastMetadata string `json:"MediaServiceFastMetadata"`
	// IsStreamClassicMigration bool `json:"isStreamClassicMigration"`
	// FirstInteractiveContentOffset int `json:"firstInteractiveContentOffset"`
	// FirstTrimmedInteractiveContentOffset int `json:"firstTrimmedInteractiveContentOffset"`

	ThumbnailURL string `json:"thumbnailUrl"` // 缩略图

	// IsAudio bool `json:"isAudio"`
	// StreamBootstrapOptions int `json:"StreamBootstrapOptions"`
	// EncKey time.Time `json:"EncKey"`
	// CdnOnly bool `json:"cdnOnly"`
	// IsPremiumVideoStreamingEnabled bool `json:"isPremiumVideoStreamingEnabled"`
	// IsHigherMeTAMaxBitrateEnabled bool `json:"isHigherMeTAMaxBitrateEnabled"`
	// IsForceAdaptiveEnabled bool `json:"isForceAdaptiveEnabled"`
	// MeTABitrateRemoveThreshold bool `json:"MeTABitrateRemoveThreshold"`
	// AreReactionsAllowed bool `json:"areReactionsAllowed"`
	// IsEndpointProtectionEnabled bool `json:"isEndpointProtectionEnabled"`
	// SpItemPreviewThumbnailsSpriteContentCdnURL string `json:"spItemPreviewThumbnailsSpriteContentCdnUrl"`
	// SpItemTimelineEventStreamCdnURL string `json:"spItemTimelineEventStreamCdnUrl"`
	// SpItemThumbnailContentCdnURL string `json:"spItemThumbnailContentCdnUrl"`
	// SpItemTranscriptsContentCdnURL string `json:"spItemTranscriptsContentCdnUrl"`
	// ProviderCdnTransformURL string `json:".providerCdnTransformUrl"`
}

// _spPageContextInfo
type shareDirInfo struct {
	WebAbsoluteURL string `json:"webAbsoluteUrl"`
	ListURL        string `json:"listUrl"`
	DriveInfo      struct {
		DriveURL         string `json:".driveUrl"`
		DriveAccessToken string `json:".driveAccessToken"`
		// DriveAccessTokenV21 string `json:".driveAccessTokenV21"`
	} `json:"driveInfo"`
}

const (
	prefixFile = `var g_fileInfo`
	prefixDir  = `var _spPageContextInfo`
)

var prefixs = [...]string{prefixFile, prefixDir}

type indexCallback func(i int, buf []byte) (bool, error)

// 修改版：找到任一特征后返回 JSON 起始段数据，在回调函数中验证后返回 true 退出
func indexStreamBufferSpecial(r io.Reader, buf []byte, cb indexCallback) (err error) {
	lseps := len(prefixs)

	// 计算最大长度
	var msep int
	for i := 0; i < lseps; i++ {
		if l := len(prefixs[i]); l > msep {
			msep = l
		}
	}

	// 检测缓冲区大小
	lbuf := len(buf)
	if msep > lbuf/2 {
		panic(`IndexStream: buffer too small!`)
	}

	// 首次读取
	var n int
	n, err = r.Read(buf)
	if err != nil {
		if err == io.EOF {
			err = nil
		} else {
			return
		}
	}

	// 首次搜索
	var pidx int = -1
	var pi int
	var ret bool
	for i := 0; i < lseps; i++ {
		idx := bytes.Index(buf[:n], util.StringToBytes(prefixs[i]))
		if idx != -1 {
			idx += len(prefixs[i])
			id2 := bytes.IndexByte(buf[idx:n], '{')
			if id2 != -1 {
				// 找到 JSON 数据开头
				if ret, err = cb(i, buf[idx+id2:n]); err != nil || ret {
					return
				}
			} else {
				// 不在缓冲区中，保留下次查询
				pidx = idx
				pi = i
			}
			// 立即返回
			break
		}
	}

	// 数据结束
	if n != lbuf {
		return
	}

	// 偏移量
	var off int
	last := lbuf - msep
	for {
		// 移动尾部数据
		copy(buf, buf[last:])
		off += last

		// 读取数据
		n, err = r.Read(buf[msep:])
		if err != nil {
			if err == io.EOF {
				err = nil
			} else {
				return
			}
		}

		// 匹配数据
		if pidx == -1 {
			for i := 0; i < lseps; i++ {
				idx := bytes.Index(buf[:msep+n], util.StringToBytes(prefixs[i]))
				if idx != -1 {
					idx += len(prefixs[i])
					id2 := bytes.IndexByte(buf[idx:msep+n], '{')
					if id2 != -1 {
						// 找到 JSON 数据开头
						if ret, err = cb(i, buf[idx+id2:msep+n]); err != nil || ret {
							return
						}
					} else {
						// 不在缓冲区中，保留下次查询
						pidx = idx + off
						pi = i
					}
					// 立即返回
					break
				}
			}
		} else {
			id2 := bytes.IndexByte(buf[:msep+n], '{')
			if id2 != -1 {
				// 找到 JSON 数据开头
				if ret, err = cb(pi, buf[id2:msep+n]); err != nil || ret {
					return
				} else {
					pidx = -1
				}
			}
		}

		// 数据结束
		if n != last {
			break
		}
	}

	return
}
