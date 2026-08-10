package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type FileInput struct {
	Path     string `json:"path" v:"required"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

type FileView struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Content  string `json:"content"`
	Order    int    `json:"order"`
}

type PasteView struct {
	ID                string     `json:"id"`
	Code              string     `json:"code"`
	ShareURL          string     `json:"shareUrl"`
	Title             string     `json:"title"`
	Description       string     `json:"description,omitempty"`
	Tags              []string   `json:"tags"`
	Files             []FileView `json:"files"`
	Visibility        string     `json:"visibility"`
	PasswordProtected bool       `json:"passwordProtected"`
	State             string     `json:"state"`
	Revision          int64      `json:"revision"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	ExpiresAt         *time.Time `json:"expiresAt,omitempty"`
}

type PasteSummaryView struct {
	Code              string     `json:"code"`
	ShareURL          string     `json:"shareUrl"`
	Title             string     `json:"title"`
	Tags              []string   `json:"tags"`
	FileCount         int        `json:"fileCount"`
	PrimaryLanguage   string     `json:"primaryLanguage"`
	Visibility        string     `json:"visibility"`
	PasswordProtected bool       `json:"passwordProtected"`
	State             string     `json:"state"`
	Revision          int64      `json:"revision"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	ExpiresAt         *time.Time `json:"expiresAt,omitempty"`
}

type CreatePasteReq struct {
	g.Meta      `path:"/api/v1/pastes" method:"post" tags:"Pastes" summary:"Create an anonymous or user-owned Paste"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Tags        []string    `json:"tags"`
	Files       []FileInput `json:"files" v:"required"`
	Visibility  string      `json:"visibility"`
	Password    string      `json:"password"`
	ExpiresAt   *time.Time  `json:"expiresAt"`
}

type CreatePasteRes struct {
	Paste PasteView `json:"paste"`
}

type GetPasteReq struct {
	g.Meta `path:"/api/v1/pastes/{code}" method:"get" tags:"Pastes" summary:"Open a Paste that does not require a password"`
	Code   string `json:"code" in:"path" v:"required"`
}

type GetPasteRes struct {
	Paste PasteView `json:"paste"`
}

type UnlockPasteReq struct {
	g.Meta   `path:"/api/v1/pastes/{code}/access" method:"post" tags:"Pastes" summary:"Open a password-protected Paste"`
	Code     string `json:"code" in:"path" v:"required"`
	Password string `json:"password" v:"required"`
}

type UnlockPasteRes struct {
	Paste PasteView `json:"paste"`
}

type ListMyPastesReq struct {
	g.Meta `path:"/api/v1/me/pastes" method:"get" tags:"My Pastes" summary:"List the current user's Pastes"`
}

type ListMyPastesRes struct {
	Pastes []PasteSummaryView `json:"pastes"`
}

type GetMyPasteReq struct {
	g.Meta `path:"/api/v1/me/pastes/{code}" method:"get" tags:"My Pastes" summary:"Get an owned Paste for management"`
	Code   string `json:"code" in:"path" v:"required"`
}

type GetMyPasteRes struct {
	Paste PasteView `json:"paste"`
}

type UpdatePasteReq struct {
	g.Meta           `path:"/api/v1/me/pastes/{code}" method:"patch" tags:"My Pastes" summary:"Update an owned Paste"`
	Code             string       `json:"code" in:"path" v:"required"`
	ExpectedRevision int64        `json:"expectedRevision" v:"required|min:1"`
	Title            *string      `json:"title"`
	Description      *string      `json:"description"`
	Tags             *[]string    `json:"tags"`
	Files            *[]FileInput `json:"files"`
	Visibility       *string      `json:"visibility"`
	Password         *string      `json:"password"`
	ExpiresAt        *time.Time   `json:"expiresAt"`
	ClearExpiry      bool         `json:"clearExpiry"`
}

type UpdatePasteRes struct {
	Paste PasteView `json:"paste"`
}

type DeletePasteReq struct {
	g.Meta           `path:"/api/v1/me/pastes/{code}" method:"delete" tags:"My Pastes" summary:"Delete an owned Paste"`
	Code             string `json:"code" in:"path" v:"required"`
	ExpectedRevision int64  `json:"expectedRevision" in:"query" v:"required|min:1"`
}

type DeletePasteRes struct{}
