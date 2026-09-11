package v1

import "github.com/gogf/gf/v2/frame/g"

type GetAuthorizationSetupReq struct {
	g.Meta `path:"/api/v1/authorization/setup" method:"get" tags:"Authorization" summary:"Read initial administrator claim status"`
}
type GetAuthorizationSetupRes struct {
	Claimed  bool `json:"claimed"`
	CanClaim bool `json:"canClaim"`
}
type ClaimAdministratorReq struct {
	g.Meta `path:"/api/v1/authorization/setup/claim" method:"post" tags:"Authorization" summary:"Claim initial administrator"`
}
type ClaimAdministratorRes struct {
	Claimed bool `json:"claimed"`
}
