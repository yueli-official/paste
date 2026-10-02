package httpapi

// PersonalAPIDeclaration exports the PAT permissions and route allowlist used
// by the running product.
func PersonalAPIDeclaration() map[string]any {
	permissions := make([]map[string]any, 0, len(personalPermissions))
	for _, permission := range personalPermissions {
		permissions = append(permissions, map[string]any{
			"key":         permission.Key,
			"label":       permission.Label,
			"description": permission.Description,
			"capability":  permission.Key,
		})
	}
	routes := make([]map[string]any, 0, len(personalRoutes))
	for _, route := range personalRoutes {
		routes = append(routes, map[string]any{
			"method":         route.method,
			"path":           route.path,
			"scopeAnyOf":     [][]string{{string(route.capability)}},
			"authentication": "pat",
		})
	}
	return map[string]any{
		"schemaVersion": 1,
		"product": map[string]string{
			"id":      "paste",
			"name":    "Paste",
			"guide":   "../developer-tokens.md",
			"openapi": "../../contracts/openapi/paste.json",
		},
		"permissions": permissions, "routes": routes, "mediaProfiles": []map[string]any{},
	}
}
