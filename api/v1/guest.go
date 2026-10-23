package v1

import "github.com/gogf/gf/v2/frame/g"

type ClaimGuestPastesReq struct {
	g.Meta `path:"/api/v1/pastes/guest-claims" method:"post" tags:"My Pastes" summary:"Claim browser guest content with an Identity guest claim token"`
}

type ClaimGuestPastesRes struct {
	Claimed int64 `json:"claimed"`
}
