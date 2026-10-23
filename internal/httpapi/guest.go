package httpapi

import (
	"context"
	"strings"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	v1 "github.com/yueli-official/paste/api/v1"
	"github.com/yueli-official/paste/internal/paste"
	"github.com/yueli-official/paste/internal/pasteerr"
)

func optionalOwner(ctx context.Context) (paste.Owner, error) {
	principal, ok := foundationauth.FromContext(ctx)
	if !ok || principal == nil {
		return paste.Owner{}, nil
	}
	if principal.HasScope("guest:claim") || strings.TrimSpace(principal.Subject) == "" {
		return paste.Owner{}, mustProblem(pasteerr.Forbidden)
	}
	if principal.IsUser() {
		return paste.Owner{UserKey: principal.Subject}, nil
	}
	if principal.SubjectKind == foundationauth.SubjectGuest {
		return paste.Owner{GuestID: principal.Subject}, nil
	}
	return paste.Owner{}, mustProblem(pasteerr.Forbidden)
}

func requiredOwner(ctx context.Context) (paste.Owner, error) {
	owner, err := optionalOwner(ctx)
	if err != nil {
		return owner, err
	}
	if !owner.Valid() {
		return owner, mustProblem(pasteerr.Unauthorized)
	}
	return owner, nil
}

func (controller *Managed) ClaimGuestPastes(ctx context.Context, _ *v1.ClaimGuestPastesReq) (*v1.ClaimGuestPastesRes, error) {
	principal, ok := foundationauth.FromContext(ctx)
	if !ok || principal == nil || !principal.IsUser() || !principal.HasScope("guest:claim") {
		return nil, mustProblem(pasteerr.Forbidden)
	}
	claim, _ := principal.Claim("guest_subject")
	guestID, _ := claim.(string)
	if strings.TrimSpace(guestID) == "" || strings.TrimSpace(principal.Subject) == "" {
		return nil, mustProblem(pasteerr.Forbidden)
	}
	count, err := controller.core.pastes.ClaimGuest(ctx, guestID, principal.Subject)
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.ClaimGuestPastesRes{Claimed: count}, nil
}
