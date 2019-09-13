package smms

import (
	"io"
	"time"
)

type smmsResponse struct {
	Success   bool        `json:"success,omitempty"`
	Code      string      `json:"code,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"RequestId,omitempty"`
}

type tokenData struct {
	Token string `json:"token"`
}

type UploadReq struct {
	Filename string
	Picture  io.Reader
}

type ImageRsp struct {
	FileID    int    `json:"file_id,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Filename  string `json:"filename,omitempty"`
	StoreName string `json:"storename,omitempty"`
	Size      int    `json:"size,omitempty"`
	Path      string `json:"path,omitempty"`
	Hash      string `json:"hash,omitempty"`
	URL       string `json:"url,omitempty"`
	Delete    string `json:"delete,omitempty"`
	Page      string `json:"page,omitempty"`
}

type ProfileRsp struct {
	Username    string    `json:"username,omitempty"`
	Role        string    `json:"role,omitempty"`
	GroupExpire time.Time `json:"group_expire,omitempty"`
	DiskUsage   string    `json:"disk_usage,omitempty"`
	DiskLimit   string    `json:"disk_limit,omitempty"`
}
