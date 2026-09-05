package paste

import "time"

const (
	MaxFiles        = 20
	MaxFileBytes    = 1 << 20
	MaxContentBytes = 1 << 20
	MaxTitleRunes   = 120
	MaxDescription  = 2000
	MaxTags         = 8
	MaxTagRunes     = 32
	MaxPathRunes    = 180
)

type Visibility string

const (
	VisibilityUnlisted Visibility = "unlisted"
	VisibilityPrivate  Visibility = "private"
)

type State string

const (
	StateActive  State = "active"
	StateDeleted State = "deleted"
)

type File struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Content  string `json:"content"`
	Order    int    `json:"order"`
}

type Paste struct {
	ID            string     `json:"id"`
	Code          string     `json:"code"`
	OwnerUserKey  string     `json:"ownerUserKey,omitempty"`
	Title         string     `json:"title"`
	Description   string     `json:"description,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	Files         []File     `json:"files"`
	Visibility    Visibility `json:"visibility"`
	PasswordHash  []byte     `json:"-"`
	State         State      `json:"state"`
	Revision      int64      `json:"revision"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	DeletedAt     *time.Time `json:"deletedAt,omitempty"`
	PasswordGuard bool       `json:"passwordProtected"`
}

func (value Paste) Anonymous() bool { return value.OwnerUserKey == "" }

type CreateInput struct {
	OwnerUserKey string
	Title        string
	Description  string
	Tags         []string
	Files        []File
	Visibility   Visibility
	Password     string
	ExpiresAt    *time.Time
}

type Access struct {
	UserKey  string
	Password string
}

type UpdateInput struct {
	OwnerUserKey     string
	ExpectedRevision int64
	Title            *string
	Description      *string
	Tags             *[]string
	Files            *[]File
	Visibility       *Visibility
	Password         *string
	ExpiresAt        *time.Time
	ClearExpiry      bool
}

type AdministrationQuery struct {
	OwnerUserKey string
	Query        string
	Visibility   Visibility
	State        State
	Ownership    string
	Limit        int
	Offset       int
}

type AdministrationItem struct {
	Code              string
	OwnerUserKey      string
	Title             string
	Tags              []string
	FileCount         int
	PrimaryLanguage   string
	Visibility        Visibility
	PasswordProtected bool
	State             State
	Revision          int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ExpiresAt         *time.Time
}

type AdministrationPage struct {
	Items  []AdministrationItem
	Total  int
	Limit  int
	Offset int
}

type GovernanceInput struct {
	ExpectedRevision int64
	Visibility       *Visibility
	ExpiresAt        *time.Time
	ClearExpiry      bool
}
