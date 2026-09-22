package pasteauthz

import "github.com/yueli-official/foundation/go/authorization"

const (
	RootScopeID              authorization.ScopeID       = "paste"
	ScopeSite                authorization.ScopeType     = "site"
	RoleAdministrator        authorization.RoleKey       = "administrator"
	CapabilityPublicRead     authorization.CapabilityKey = "paste.public.read"
	CapabilityContentManage  authorization.CapabilityKey = "paste.content.manage"
	CapabilityPasteCreate    authorization.CapabilityKey = "paste.paste.create"
	CapabilityPasteRead      authorization.CapabilityKey = "paste.paste.read"
	CapabilityPasteUpdate    authorization.CapabilityKey = "paste.paste.update"
	CapabilityPasteDelete    authorization.CapabilityKey = "paste.paste.delete"
	CapabilityPasteModerate  authorization.CapabilityKey = "paste.paste.moderate"
	CapabilitySettingsManage authorization.CapabilityKey = "paste.settings.manage"
)

func Definition() authorization.Definition {
	return authorization.Definition{
		Consumer: "paste",
		Version:  2,
		Capabilities: []authorization.CapabilityDefinition{
			{
				Key: CapabilityPublicRead, Version: 1,
				Binding:       authorization.BindingAccessLayerEligible,
				AllowedScopes: []authorization.ScopeType{ScopeSite},
			},
			{
				Key: CapabilityPasteCreate, Version: 1,
				Binding:       authorization.BindingAccessLayerEligible,
				AllowedScopes: []authorization.ScopeType{ScopeSite},
			},
			{
				Key: CapabilityPasteRead, Version: 1,
				Binding:       authorization.BindingAccessLayerEligible,
				AllowedScopes: []authorization.ScopeType{ScopeSite},
			},
			{
				Key: CapabilityPasteUpdate, Version: 1,
				Binding:       authorization.BindingAccessLayerEligible,
				AllowedScopes: []authorization.ScopeType{ScopeSite},
			},
			{
				Key: CapabilityPasteDelete, Version: 1,
				Binding:       authorization.BindingAccessLayerEligible,
				AllowedScopes: []authorization.ScopeType{ScopeSite},
			},
			{
				Key:              CapabilityPasteModerate,
				Version:          1,
				Binding:          authorization.BindingProtectedOnly,
				Risk:             authorization.RiskHigh,
				Audit:            authorization.AuditFull,
				AllowedScopes:    []authorization.ScopeType{ScopeSite},
				EligibleSubjects: []authorization.SubjectKind{authorization.SubjectUser},
			},
			{
				Key:              CapabilitySettingsManage,
				Version:          1,
				Binding:          authorization.BindingProtectedOnly,
				Risk:             authorization.RiskHigh,
				Audit:            authorization.AuditFull,
				AllowedScopes:    []authorization.ScopeType{ScopeSite},
				EligibleSubjects: []authorization.SubjectKind{authorization.SubjectUser},
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
				CapabilityPasteCreate,
				CapabilityPasteRead,
				CapabilityPasteUpdate,
				CapabilityPasteDelete,
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
				CapabilityPasteModerate,
				CapabilitySettingsManage,
			},
		}},
	}
}
