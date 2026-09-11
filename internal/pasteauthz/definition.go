package pasteauthz

import "github.com/yueli-official/foundation/go/authorization"

const (
	RootScopeID             authorization.ScopeID       = "paste"
	ScopeSite               authorization.ScopeType     = "site"
	RoleAdministrator       authorization.RoleKey       = "administrator"
	CapabilityPublicRead    authorization.CapabilityKey = "paste.public.read"
	CapabilityContentManage authorization.CapabilityKey = "paste.content.manage"
)

func Definition() authorization.Definition {
	return authorization.Definition{
		Consumer: "paste",
		Version:  1,
		Capabilities: []authorization.CapabilityDefinition{
			{
				Key: CapabilityPublicRead, Version: 1,
				Binding:       authorization.BindingAccessLayerEligible,
				AllowedScopes: []authorization.ScopeType{ScopeSite},
			},
			{
				Key:              CapabilityContentManage,
				Version:          1,
				Binding:          authorization.BindingProtectedOnly,
				Risk:             authorization.RiskHigh,
				Audit:            authorization.AuditFull,
				AllowedScopes:    []authorization.ScopeType{ScopeSite},
				EligibleSubjects: []authorization.SubjectKind{authorization.SubjectUser},
			},
		},
		Scopes: authorization.ScopeSchema{Types: []authorization.ScopeTypeDefinition{{
			Key: ScopeSite, Root: true,
		}}},
		AccessLayers: []authorization.AccessLayerDefinition{
			{Key: authorization.AccessLayerVisitor, Capabilities: []authorization.CapabilityKey{CapabilityPublicRead}},
			{Key: authorization.AccessLayerAuthenticated, Capabilities: []authorization.CapabilityKey{
				authorization.CapabilityApplicationCreate,
				authorization.CapabilityApplicationReadOwn,
				authorization.CapabilityApplicationWithdraw,
				authorization.CapabilityInvitationAccept,
			}},
		},
		Roles: []authorization.RoleDefinition{{
			Key:         RoleAdministrator,
			DisplayName: "管理员",
			Protected:   true,
			Capabilities: []authorization.CapabilityKey{
				authorization.CapabilityManage,
				authorization.CapabilityAuditRead,
				CapabilityContentManage,
			},
		}},
	}
}
