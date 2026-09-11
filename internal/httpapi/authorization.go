package httpapi

import (
	"context"
	v1 "github.com/yueli-official/paste/api/v1"
	"github.com/yueli-official/paste/internal/pasteerr"
)

func (controller *Public) GetAuthorizationSetup(ctx context.Context, _ *v1.GetAuthorizationSetupReq) (*v1.GetAuthorizationSetupRes, error) {
	status, err := controller.core.authorization.AdministratorClaimStatus(ctx)
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.GetAuthorizationSetupRes{Claimed: status.Claimed, CanClaim: !status.Claimed}, nil
}
func (controller *Managed) ClaimAdministrator(ctx context.Context, _ *v1.ClaimAdministratorReq) (*v1.ClaimAdministratorRes, error) {
	if _, err := requiredUser(ctx); err != nil {
		return nil, err
	}
	result, err := controller.core.authorization.ClaimInitialAdministrator(ctx)
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.ClaimAdministratorRes{Claimed: result.Status.Claimed}, nil
}
