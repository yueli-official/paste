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

type AdministrationPasteView struct {
	Code              string     `json:"code"`
	ShareURL          string     `json:"shareUrl"`
	OwnerUserKey      string     `json:"ownerUserKey,omitempty"`
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

type SiteSettingsView struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Revision    int64     `json:"revision"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type GovernanceSettingsView struct {
	UserDailyLimit      int       `json:"userDailyLimit"`
	AnonymousDailyLimit int       `json:"anonymousDailyLimit"`
	Revision            int64     `json:"revision"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type AdministrationUserView struct {
	UserKey             string     `json:"userKey"`
	State               string     `json:"state"`
	DailyLimitOverride  *int       `json:"dailyLimitOverride,omitempty"`
	EffectiveDailyLimit int        `json:"effectiveDailyLimit"`
	UsedToday           int        `json:"usedToday"`
	TotalPastes         int        `json:"totalPastes"`
	ActivePastes        int        `json:"activePastes"`
	LastCreatedAt       *time.Time `json:"lastCreatedAt,omitempty"`
	Reason              string     `json:"reason,omitempty"`
	Revision            int64      `json:"revision"`
	UpdatedAt           *time.Time `json:"updatedAt,omitempty"`
}

type AdministrationUserPolicyView struct {
	UserKey            string    `json:"userKey"`
	State              string    `json:"state"`
	DailyLimitOverride *int      `json:"dailyLimitOverride,omitempty"`
	Reason             string    `json:"reason,omitempty"`
	Revision           int64     `json:"revision"`
	UpdatedAt          time.Time `json:"updatedAt"`
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

type GetSiteSettingsReq struct {
	g.Meta `path:"/api/v1/settings" method:"get" tags:"Site" summary:"Read public site settings"`
}

type GetSiteSettingsRes struct {
	Settings SiteSettingsView `json:"settings"`
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

type ListAdministrationPastesReq struct {
	g.Meta     `path:"/api/v1/admin/pastes" method:"get" tags:"Administration" summary:"List Pastes for site governance"`
	Query      string `json:"q" in:"query"`
	Visibility string `json:"visibility" in:"query"`
	State      string `json:"state" in:"query"`
	Ownership  string `json:"ownership" in:"query"`
	Limit      int    `json:"limit" in:"query"`
	Offset     int    `json:"offset" in:"query"`
}

type GetAdministrationSessionReq struct {
	g.Meta `path:"/api/v1/admin/session" method:"get" tags:"Administration" summary:"Check Paste administrator access"`
}

type GetAdministrationSessionRes struct {
	Allowed bool `json:"allowed"`
	UserKey string `json:"userKey"`
}

type ListAdministrationPastesRes struct {
	Pastes []AdministrationPasteView `json:"pastes"`
	Total  int                       `json:"total"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
}

type GovernPasteReq struct {
	g.Meta           `path:"/api/v1/admin/pastes/{code}" method:"patch" tags:"Administration" summary:"Govern a Paste"`
	Code             string     `json:"code" in:"path" v:"required"`
	ExpectedRevision int64      `json:"expectedRevision" v:"required|min:1"`
	Visibility       *string    `json:"visibility"`
	ExpiresAt        *time.Time `json:"expiresAt"`
	ClearExpiry      bool       `json:"clearExpiry"`
}

type GovernPasteRes struct {
	Paste AdministrationPasteView `json:"paste"`
}

type AdministrationDeletePasteReq struct {
	g.Meta           `path:"/api/v1/admin/pastes/{code}" method:"delete" tags:"Administration" summary:"Delete a Paste as site administrator"`
	Code             string `json:"code" in:"path" v:"required"`
	ExpectedRevision int64  `json:"expectedRevision" in:"query" v:"required|min:1"`
}

type AdministrationDeletePasteRes struct{}

type ListAdministrationUsersReq struct {
	g.Meta `path:"/api/v1/admin/users" method:"get" tags:"Administration" summary:"List Paste users for product governance"`
	Query  string `json:"q" in:"query"`
	State  string `json:"state" in:"query"`
	Limit  int    `json:"limit" in:"query"`
	Offset int    `json:"offset" in:"query"`
}

type ListAdministrationUsersRes struct {
	Users  []AdministrationUserView `json:"users"`
	Total  int                      `json:"total"`
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
}

type UpdateAdministrationUserReq struct {
	g.Meta             `path:"/api/v1/admin/users/{userKey}" method:"patch" tags:"Administration" summary:"Update a Paste user creation policy"`
	UserKey            string `json:"userKey" in:"path" v:"required"`
	State              string `json:"state"`
	DailyLimitOverride *int   `json:"dailyLimitOverride"`
	ClearDailyLimit    bool   `json:"clearDailyLimit"`
	Reason             string `json:"reason"`
	ExpectedRevision   int64  `json:"expectedRevision" v:"min:0"`
}

type UpdateAdministrationUserRes struct {
	User AdministrationUserPolicyView `json:"user"`
}

type GetGovernanceSettingsReq struct {
	g.Meta `path:"/api/v1/admin/governance-settings" method:"get" tags:"Administration" summary:"Read Paste abuse-control settings"`
}

type GetGovernanceSettingsRes struct {
	Settings GovernanceSettingsView `json:"settings"`
}

type UpdateGovernanceSettingsReq struct {
	g.Meta              `path:"/api/v1/admin/governance-settings" method:"patch" tags:"Administration" summary:"Update Paste abuse-control settings"`
	UserDailyLimit      int   `json:"userDailyLimit"`
	AnonymousDailyLimit int   `json:"anonymousDailyLimit"`
	ExpectedRevision    int64 `json:"expectedRevision" v:"required|min:1"`
}

type UpdateGovernanceSettingsRes struct {
	Settings GovernanceSettingsView `json:"settings"`
}

type UpdateSiteSettingsReq struct {
	g.Meta           `path:"/api/v1/admin/settings" method:"patch" tags:"Administration" summary:"Update public site settings"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	ExpectedRevision int64  `json:"expectedRevision" v:"required|min:1"`
}

type UpdateSiteSettingsRes struct {
	Settings SiteSettingsView `json:"settings"`
}
