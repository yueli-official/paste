package site

import "time"

const (
	DefaultName         = "代码片段"
	DefaultDescription  = "轻量、专注的多文件代码分享。"
	MaxNameRunes        = 40
	MaxDescriptionRunes = 160
)

type Settings struct {
	Name        string
	Description string
	Revision    int64
	UpdatedAt   time.Time
	UpdatedBy   string
}

type UpdateInput struct {
	Name             string
	Description      string
	ExpectedRevision int64
	UpdatedBy        string
}
