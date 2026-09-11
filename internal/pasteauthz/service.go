package pasteauthz

import (
	"context"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	"github.com/yueli-official/paste/internal/paste"
)

type Runtime interface {
	authorization.Authorizer
	authorization.AccessReader
	authorization.AdministratorClaimer
}

type Service struct{ runtime Runtime }

func New(runtime Runtime) *Service { return &Service{runtime: runtime} }

func (service *Service) Subject(ctx context.Context) authorization.SubjectRef {
	principal, ok := foundationauth.FromContext(ctx)
	if !ok {
		return authorization.SubjectRef{Kind: authorization.SubjectAnonymous}
	}
	if principal.SubjectKind == foundationauth.SubjectUser && principal.Subject != "" {
		return authorization.SubjectRef{Kind: authorization.SubjectUser, ID: principal.Subject}
	}
	subjectKind, _ := principal.Claim("subject_kind")
	if subjectKind == "user" && principal.Subject != "" {
		return authorization.SubjectRef{Kind: authorization.SubjectUser, ID: principal.Subject}
	}
	return authorization.SubjectRef{Kind: authorization.SubjectAnonymous}
}

func (service *Service) AdministratorClaimStatus(ctx context.Context) (authorization.AdministratorClaimStatus, error) {
	return service.runtime.AdministratorClaimStatus(ctx)
}

func (service *Service) ClaimInitialAdministrator(ctx context.Context) (authorization.ClaimInitialAdministratorResult, error) {
	return service.runtime.ClaimInitialAdministrator(ctx, authorization.ClaimInitialAdministratorCommand{Actor: service.Subject(ctx)})
}

func (service *Service) EffectiveAccess(ctx context.Context) (authorization.EffectiveAccess, error) {
	return service.runtime.EffectiveAccess(ctx, authorization.EffectiveAccessQuery{
		Subject: service.Subject(ctx), ScopeID: RootScopeID,
	})
}

func (service *Service) RequireManage(ctx context.Context) error {
	decision, err := service.runtime.Decide(ctx, authorization.DecisionRequest{
		Subject: service.Subject(ctx), Capability: CapabilityContentManage, ScopeID: RootScopeID,
	})
	if err != nil {
		return err
	}
	if !decision.Allowed {
		return paste.ErrForbidden
	}
	return nil
}
