package share

import (
	"context"
	"time"
)

type DriveItem struct {
	ContentDownloadURL   string    `json:"@content.downloadUrl,omitempty"`
	CreatedDateTime      time.Time `json:"createdDateTime"`
	ETag                 string    `json:"eTag"`
	ID                   string    `json:"id"`
	LastModifiedDateTime time.Time `json:"lastModifiedDateTime"`
	Name                 string    `json:"name"`
	ParentReference      struct {
		// DriveType string `json:"driveType"`

		DriveID string `json:"driveId"`
		// ID string `json:"id"`

		Path string `json:"path"` // "/drives/{DriveID}/root:" 一般长度为 30 或 80
	} `json:"parentReference"`
	CTag string `json:"cTag"`
	File *struct {
		FileExtension string `json:"fileExtension"`
		Hashes        struct {
			QuickXorHash string `json:"quickXorHash"`
			Sha1Hash     string `json:"sha1Hash,omitempty"`
			Sha256Hash   string `json:"sha256Hash,omitempty"`
		} `json:"hashes"`
		MimeType string `json:"mimeType"`
	} `json:"file,omitempty"`
	Folder *struct {
		ChildCount int `json:"childCount"`
	} `json:"folder,omitempty"`
	Size int64 `json:"size"`
}

// https://learn.microsoft.com/zh-cn/graph/api/driveitem-list-children?view=graph-rest-1.0&tabs=http

type DriveChildren struct {
	Value []DriveItem `json:"value"`

	OdataNextLink string `json:"@odata.nextLink,omitempty"`
}

type DriveThumbs struct {
	Value []struct {
		ID    string `json:"id"`
		Large struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
		} `json:"large"`
		Medium struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
		} `json:"medium"`
		Small struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
		} `json:"small"`
	} `json:"value"`
}

type Item interface {
	GetItem() *DriveItem
	ListItem(ctx context.Context, subPath, next string) (*DriveChildren, error)
}
