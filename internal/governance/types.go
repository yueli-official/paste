package governance

import "time"

const (
	DefaultUserDailyLimit      = 50
	DefaultAnonymousDailyLimit = 200
	MaxUserDailyLimit          = 10_000
	MaxAnonymousDailyLimit     = 100_000
	MaxReasonRunes             = 240
)

type UserState string

const (
	UserStateActive    UserState = "active"
	UserStateSuspended UserState = "suspended"
)

type Settings struct {
	UserDailyLimit      int
	AnonymousDailyLimit int
	Revision            int64
	UpdatedAt           time.Time
	UpdatedBy           string
}

type UpdateSettingsInput struct {
	UserDailyLimit      int
	AnonymousDailyLimit int
	ExpectedRevision    int64
	UpdatedBy           string
}

type UserPolicy struct {
	UserKey            string
	State              UserState
	DailyLimitOverride *int
	Reason             string
	Revision           int64
	UpdatedAt          time.Time
	UpdatedBy          string
}

type UpdateUserInput struct {
	UserKey            string
	State              UserState
	DailyLimitOverride *int
	ClearDailyLimit    bool
	Reason             string
	ExpectedRevision   int64
	UpdatedBy          string
}

type UserQuery struct {
	Query  string
	State  UserState
	Limit  int
	Offset int
}

type User struct {
	UserKey             string
	State               UserState
	DailyLimitOverride  *int
	EffectiveDailyLimit int
	UsedToday           int
	TotalPastes         int
	ActivePastes        int
	LastCreatedAt       *time.Time
	Reason              string
	Revision            int64
	UpdatedAt           *time.Time
	UpdatedBy           string
}

type UserPage struct {
	Users  []User
	Total  int
	Limit  int
	Offset int
}

func EffectiveDailyLimit(settings Settings, policy UserPolicy) int {
	if policy.DailyLimitOverride != nil {
		return *policy.DailyLimitOverride
	}
	return settings.UserDailyLimit
}
