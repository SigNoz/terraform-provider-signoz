// Layer 3 expand/flatten for auth_domain — the largest conv surface in
// the provider. Five nested customtypes (AttributeMapping, GoogleConfig,
// OIDCConfig, SamlConfig, RoleMapping, AuthNProviderInfo) plus a
// wrapping AuthDomainConfig that recombines four of them into a
// oneOf-shaped block.
//
// Resource and datasource models are SIBLINGS (neither is a strict
// subset of the other) — resource has Config, datasource has CreatedAt
// / UpdatedAt. `AuthDomainFlat` is the union of both with
// `ToResource` / `ToDataSource` narrowing methods (Pattern C —
// wide-flat-narrow); the codegen `services` template emits
// `next.ToResource()` / `next.ToDataSource()` accordingly.
package conv

import (
	"context"

	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
	"github.com/SigNoz/terraform-provider-signoz/internal/schemas"
	customtypes "github.com/SigNoz/terraform-provider-signoz/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ---------------------------------------------------------------------------
// Layer 3 — resource expanders/flatteners
// ---------------------------------------------------------------------------

// ExpandAuthtypesPostableAuthDomain converts the framework resource model into
// the POST body. Reads the nested config customtype and forwards it.
func ExpandAuthtypesPostableAuthDomain(ctx context.Context, m schemas.AuthDomainModel) (*apitypes.AuthtypesPostableAuthDomain, diag.Diagnostics) {
	cfg, diags := ExpandAuthDomainConfig(ctx, m.Config)
	if diags.HasError() {
		return nil, diags
	}
	name := m.Name.ValueString()
	return &apitypes.AuthtypesPostableAuthDomain{
		Name:   &name,
		Config: cfg,
	}, diags
}

// ExpandAuthtypesUpdateableAuthDomain is the PUT-shape variant. The API rejects
// `name` on update — only Config travels.
func ExpandAuthtypesUpdateableAuthDomain(ctx context.Context, m schemas.AuthDomainModel) (*apitypes.AuthtypesUpdateableAuthDomain, diag.Diagnostics) {
	cfg, diags := ExpandAuthDomainConfig(ctx, m.Config)
	if diags.HasError() {
		return nil, diags
	}
	return &apitypes.AuthtypesUpdateableAuthDomain{Config: cfg}, diags
}

// AuthDomainFlat is the union of fields needed to populate either
// AuthDomainModel (resource — has Config, no audit) or
// AuthDomainDataSourceModel (datasource — no Config, has audit).
type AuthDomainFlat struct {
	AuthNproviderInfo customtypes.AuthtypesAuthNproviderInfoValue
	Config            customtypes.AuthtypesAuthDomainConfigValue
	CreatedAt         types.String
	GoogleAuthConfig  customtypes.AuthtypesGoogleConfigValue
	ID                types.String
	Name              types.String
	OidcConfig        customtypes.AuthtypesOidcconfigValue
	OrgID             types.String
	RoleMapping       customtypes.AuthtypesRoleMappingValue
	SamlConfig        customtypes.AuthtypesSamlConfigValue
	SsoEnabled        types.Bool
	SsoType           types.String
	UpdatedAt         types.String
}

// ToResource narrows the flat shape down to the resource model.
func (f *AuthDomainFlat) ToResource() *schemas.AuthDomainModel {
	return &schemas.AuthDomainModel{
		AuthNproviderInfo: f.AuthNproviderInfo,
		Config:            f.Config,
		GoogleAuthConfig:  f.GoogleAuthConfig,
		Id:                f.ID,
		Name:              f.Name,
		OidcConfig:        f.OidcConfig,
		OrgId:             f.OrgID,
		RoleMapping:       f.RoleMapping,
		SamlConfig:        f.SamlConfig,
		SsoEnabled:        f.SsoEnabled,
		SsoType:           f.SsoType,
	}
}

// ToDataSource narrows the flat shape down to the datasource model.
func (f *AuthDomainFlat) ToDataSource() *schemas.AuthDomainDataSourceModel {
	return &schemas.AuthDomainDataSourceModel{
		AuthNproviderInfo: f.AuthNproviderInfo,
		CreatedAt:         f.CreatedAt,
		GoogleAuthConfig:  f.GoogleAuthConfig,
		Id:                f.ID,
		Name:              f.Name,
		OidcConfig:        f.OidcConfig,
		OrgId:             f.OrgID,
		RoleMapping:       f.RoleMapping,
		SamlConfig:        f.SamlConfig,
		SsoEnabled:        f.SsoEnabled,
		SsoType:           f.SsoType,
		UpdatedAt:         f.UpdatedAt,
	}
}

// FlattenAuthtypesGettableAuthDomain converts a server response into the flat shape. The
// resource and datasource pick their respective fields via the To* helpers.
func FlattenAuthtypesGettableAuthDomain(ctx context.Context, g *apitypes.AuthtypesGettableAuthDomain) (*AuthDomainFlat, diag.Diagnostics) {
	var diags diag.Diagnostics
	if g == nil {
		return nil, diags
	}

	googleObj, d := flattenGoogleConfigInner(ctx, g.GoogleAuthConfig)
	diags.Append(d...)
	oidcObj, d := flattenOIDCConfigInner(ctx, g.OidcConfig)
	diags.Append(d...)
	samlObj, d := flattenSamlConfigInner(ctx, g.SamlConfig)
	diags.Append(d...)
	roleObj, d := flattenRoleMappingInner(ctx, g.RoleMapping)
	diags.Append(d...)
	authnInfoObj, d := flattenAuthNProviderInfoInner(ctx, g.AuthNProviderInfo)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	googleTyped, d := castToGoogleConfigTyped(ctx, googleObj)
	diags.Append(d...)
	oidcTyped, d := castToOIDCConfigTyped(ctx, oidcObj)
	diags.Append(d...)
	samlTyped, d := castToSamlConfigTyped(ctx, samlObj)
	diags.Append(d...)
	roleTyped, d := castToRoleMappingTyped(ctx, roleObj)
	diags.Append(d...)
	authnInfoTyped, d := castToAuthNProviderInfoTyped(ctx, authnInfoObj)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	var ssoType types.String
	if g.SsoType != nil {
		ssoType = types.StringValue(string(*g.SsoType))
	} else {
		ssoType = types.StringNull()
	}
	var ssoEnabled types.Bool
	if g.SsoEnabled != nil {
		ssoEnabled = types.BoolValue(*g.SsoEnabled)
	} else {
		ssoEnabled = types.BoolNull()
	}

	configTyped, d := buildAuthDomainConfigValue(ctx,
		googleObj, oidcObj, samlObj, roleObj,
		ssoEnabled, ssoType,
	)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &AuthDomainFlat{
		AuthNproviderInfo: authnInfoTyped,
		Config:            configTyped,
		CreatedAt:         TimeStringFromPointer(g.CreatedAt),
		GoogleAuthConfig:  googleTyped,
		ID:                types.StringValue(g.Id),
		Name:              StringFromPointer(g.Name),
		OidcConfig:        oidcTyped,
		OrgID:             StringFromPointer(g.OrgId),
		RoleMapping:       roleTyped,
		SamlConfig:        samlTyped,
		SsoEnabled:        ssoEnabled,
		SsoType:           ssoType,
		UpdatedAt:         TimeStringFromPointer(g.UpdatedAt),
	}, diags
}

// ---------------------------------------------------------------------------
// Layer 2 — per-customtype expand/flatten
// ---------------------------------------------------------------------------

// ExpandAuthDomainConfig walks the AuthtypesAuthDomainConfigValue (the
// nested `config` block in the schema). Reads the three oneOf children
// plus the independent role_mapping / sso_* fields.
func ExpandAuthDomainConfig(ctx context.Context, sv customtypes.AuthtypesAuthDomainConfigValue) (*apitypes.AuthtypesAuthDomainConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if sv.IsNull() || sv.IsUnknown() {
		return nil, diags
	}

	google, d := expandGoogleConfig(ctx, sv.GoogleAuthConfig)
	diags.Append(d...)
	oidc, d := expandOIDCConfig(ctx, sv.OidcConfig)
	diags.Append(d...)
	saml, d := expandSamlConfig(ctx, sv.SamlConfig)
	diags.Append(d...)
	role, d := expandRoleMapping(ctx, sv.RoleMapping)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	out := &apitypes.AuthtypesAuthDomainConfig{
		GoogleAuthConfig: google,
		OidcConfig:       oidc,
		SamlConfig:       saml,
		RoleMapping:      role,
		SsoEnabled:       BoolPointer(sv.SsoEnabled),
	}
	if !sv.SsoType.IsNull() && !sv.SsoType.IsUnknown() {
		t := apitypes.AuthtypesAuthNProvider(sv.SsoType.ValueString())
		out.SsoType = &t
	}
	return out, diags
}

func expandGoogleConfig(ctx context.Context, ov basetypes.ObjectValue) (*apitypes.AuthtypesGoogleConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ov.IsNull() || ov.IsUnknown() {
		return nil, diags
	}
	gv, d := customtypes.NewAuthtypesGoogleConfigValue(
		customtypes.AuthtypesGoogleConfigValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	allowed, d := StringPointerSliceFromList(ctx, gv.AllowedGroups)
	diags.Append(d...)
	domains, d := StringMapPointerFromMap(ctx, gv.DomainToAdminEmail)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &apitypes.AuthtypesGoogleConfig{
		AllowedGroups:                  allowed,
		ClientId:                       StringPointer(gv.ClientId),
		ClientSecret:                   StringPointer(gv.ClientSecret),
		DomainToAdminEmail:             domains,
		FetchGroups:                    BoolPointer(gv.FetchGroups),
		FetchTransitiveGroupMembership: BoolPointer(gv.FetchTransitiveGroupMembership),
		InsecureSkipEmailVerified:      BoolPointer(gv.InsecureSkipEmailVerified),
		RedirectURI:                    StringPointer(gv.RedirectUri),
		ServiceAccountJson:             StringPointer(gv.ServiceAccountJson),
	}, diags
}

func expandOIDCConfig(ctx context.Context, ov basetypes.ObjectValue) (*apitypes.AuthtypesOIDCConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ov.IsNull() || ov.IsUnknown() {
		return nil, diags
	}
	ov2, d := customtypes.NewAuthtypesOidcconfigValue(
		customtypes.AuthtypesOidcconfigValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	claim, d := expandAttributeMapping(ctx, ov2.ClaimMapping)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &apitypes.AuthtypesOIDCConfig{
		ClaimMapping:              claim,
		ClientId:                  StringPointer(ov2.ClientId),
		ClientSecret:              StringPointer(ov2.ClientSecret),
		GetUserInfo:               BoolPointer(ov2.GetUserInfo),
		InsecureSkipEmailVerified: BoolPointer(ov2.InsecureSkipEmailVerified),
		Issuer:                    StringPointer(ov2.Issuer),
		IssuerAlias:               StringPointer(ov2.IssuerAlias),
	}, diags
}

func expandSamlConfig(ctx context.Context, ov basetypes.ObjectValue) (*apitypes.AuthtypesSamlConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ov.IsNull() || ov.IsUnknown() {
		return nil, diags
	}
	ov2, d := customtypes.NewAuthtypesSamlConfigValue(
		customtypes.AuthtypesSamlConfigValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	mapping, d := expandAttributeMapping(ctx, ov2.AttributeMapping)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &apitypes.AuthtypesSamlConfig{
		AttributeMapping:                mapping,
		InsecureSkipAuthNRequestsSigned: BoolPointer(ov2.InsecureSkipAuthNrequestsSigned),
		SamlCert:                        StringPointer(ov2.SamlCert),
		SamlEntity:                      StringPointer(ov2.SamlEntity),
		SamlIdp:                         StringPointer(ov2.SamlIdp),
	}, diags
}

func expandAttributeMapping(ctx context.Context, ov basetypes.ObjectValue) (*apitypes.AuthtypesAttributeMapping, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ov.IsNull() || ov.IsUnknown() {
		return nil, diags
	}
	av, d := customtypes.NewAuthtypesAttributeMappingValue(
		customtypes.AuthtypesAttributeMappingValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	return &apitypes.AuthtypesAttributeMapping{
		Email:  StringPointer(av.Email),
		Groups: StringPointer(av.Groups),
		Name:   StringPointer(av.Name),
		Role:   StringPointer(av.Role),
	}, diags
}

func expandRoleMapping(ctx context.Context, ov basetypes.ObjectValue) (*apitypes.AuthtypesRoleMapping, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ov.IsNull() || ov.IsUnknown() {
		return nil, diags
	}
	rv, d := customtypes.NewAuthtypesRoleMappingValue(
		customtypes.AuthtypesRoleMappingValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	groups, d := StringMapPointerFromMap(ctx, rv.GroupMappings)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	return &apitypes.AuthtypesRoleMapping{
		DefaultRole:      StringPointer(rv.DefaultRole),
		GroupMappings:    groups,
		UseRoleAttribute: BoolPointer(rv.UseRoleAttribute),
	}, diags
}

// ---------------------------------------------------------------------------
// Inner ObjectValue builders (flatten)
// ---------------------------------------------------------------------------

func flattenGoogleConfigInner(ctx context.Context, c *apitypes.AuthtypesGoogleConfig) (basetypes.ObjectValue, diag.Diagnostics) {
	attrTypes := customtypes.AuthtypesGoogleConfigValue{}.AttributeTypes(ctx)
	if c == nil {
		return types.ObjectNull(attrTypes), nil
	}
	var diags diag.Diagnostics
	allowed, d := ListFromStringPointerSlice(ctx, c.AllowedGroups)
	diags.Append(d...)
	domains, d := MapFromStringPointerMap(ctx, c.DomainToAdminEmail)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectUnknown(attrTypes), diags
	}
	obj, d := types.ObjectValue(attrTypes, map[string]attr.Value{
		"allowed_groups":                    allowed,
		"client_id":                         StringFromPointer(c.ClientId),
		"client_secret":                     StringFromPointer(c.ClientSecret),
		"domain_to_admin_email":             domains,
		"fetch_groups":                      BoolFromPointer(c.FetchGroups),
		"fetch_transitive_group_membership": BoolFromPointer(c.FetchTransitiveGroupMembership),
		"insecure_skip_email_verified":      BoolFromPointer(c.InsecureSkipEmailVerified),
		"redirect_uri":                      StringFromPointer(c.RedirectURI),
		"service_account_json":              StringFromPointer(c.ServiceAccountJson),
	})
	diags.Append(d...)
	return obj, diags
}

func flattenOIDCConfigInner(ctx context.Context, c *apitypes.AuthtypesOIDCConfig) (basetypes.ObjectValue, diag.Diagnostics) {
	attrTypes := customtypes.AuthtypesOidcconfigValue{}.AttributeTypes(ctx)
	if c == nil {
		return types.ObjectNull(attrTypes), nil
	}
	var diags diag.Diagnostics
	claim, d := flattenAttributeMappingInner(ctx, c.ClaimMapping)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectUnknown(attrTypes), diags
	}
	obj, d := types.ObjectValue(attrTypes, map[string]attr.Value{
		"claim_mapping":                claim,
		"client_id":                    StringFromPointer(c.ClientId),
		"client_secret":                StringFromPointer(c.ClientSecret),
		"get_user_info":                BoolFromPointer(c.GetUserInfo),
		"insecure_skip_email_verified": BoolFromPointer(c.InsecureSkipEmailVerified),
		"issuer":                       StringFromPointer(c.Issuer),
		"issuer_alias":                 StringFromPointer(c.IssuerAlias),
	})
	diags.Append(d...)
	return obj, diags
}

func flattenSamlConfigInner(ctx context.Context, c *apitypes.AuthtypesSamlConfig) (basetypes.ObjectValue, diag.Diagnostics) {
	attrTypes := customtypes.AuthtypesSamlConfigValue{}.AttributeTypes(ctx)
	if c == nil {
		return types.ObjectNull(attrTypes), nil
	}
	var diags diag.Diagnostics
	mapping, d := flattenAttributeMappingInner(ctx, c.AttributeMapping)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectUnknown(attrTypes), diags
	}
	obj, d := types.ObjectValue(attrTypes, map[string]attr.Value{
		"attribute_mapping":                   mapping,
		"insecure_skip_auth_nrequests_signed": BoolFromPointer(c.InsecureSkipAuthNRequestsSigned),
		"saml_cert":                           StringFromPointer(c.SamlCert),
		"saml_entity":                         StringFromPointer(c.SamlEntity),
		"saml_idp":                            StringFromPointer(c.SamlIdp),
	})
	diags.Append(d...)
	return obj, diags
}

func flattenAttributeMappingInner(ctx context.Context, c *apitypes.AuthtypesAttributeMapping) (basetypes.ObjectValue, diag.Diagnostics) {
	attrTypes := customtypes.AuthtypesAttributeMappingValue{}.AttributeTypes(ctx)
	if c == nil {
		return types.ObjectNull(attrTypes), nil
	}
	obj, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
		"email":  StringFromPointer(c.Email),
		"groups": StringFromPointer(c.Groups),
		"name":   StringFromPointer(c.Name),
		"role":   StringFromPointer(c.Role),
	})
	return obj, diags
}

func flattenRoleMappingInner(ctx context.Context, c *apitypes.AuthtypesRoleMapping) (basetypes.ObjectValue, diag.Diagnostics) {
	attrTypes := customtypes.AuthtypesRoleMappingValue{}.AttributeTypes(ctx)
	if c == nil {
		return types.ObjectNull(attrTypes), nil
	}
	var diags diag.Diagnostics
	groups, d := MapFromStringPointerMap(ctx, c.GroupMappings)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectUnknown(attrTypes), diags
	}
	obj, d := types.ObjectValue(attrTypes, map[string]attr.Value{
		"default_role":       StringFromPointer(c.DefaultRole),
		"group_mappings":     groups,
		"use_role_attribute": BoolFromPointer(c.UseRoleAttribute),
	})
	diags.Append(d...)
	return obj, diags
}

func flattenAuthNProviderInfoInner(ctx context.Context, c *apitypes.AuthtypesAuthNProviderInfo) (basetypes.ObjectValue, diag.Diagnostics) {
	attrTypes := customtypes.AuthtypesAuthNproviderInfoValue{}.AttributeTypes(ctx)
	if c == nil {
		return types.ObjectNull(attrTypes), nil
	}
	obj, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
		"relay_state_path": StringFromPointer(c.RelayStatePath),
	})
	return obj, diags
}

// ---------------------------------------------------------------------------
// Inner ObjectValue → typed customtype value casts
// ---------------------------------------------------------------------------

func castToGoogleConfigTyped(ctx context.Context, ov basetypes.ObjectValue) (customtypes.AuthtypesGoogleConfigValue, diag.Diagnostics) {
	if ov.IsNull() {
		return customtypes.NewAuthtypesGoogleConfigValueNull(), nil
	}
	if ov.IsUnknown() {
		return customtypes.NewAuthtypesGoogleConfigValueUnknown(), nil
	}
	return customtypes.NewAuthtypesGoogleConfigValue(
		customtypes.AuthtypesGoogleConfigValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
}

func castToOIDCConfigTyped(ctx context.Context, ov basetypes.ObjectValue) (customtypes.AuthtypesOidcconfigValue, diag.Diagnostics) {
	if ov.IsNull() {
		return customtypes.NewAuthtypesOidcconfigValueNull(), nil
	}
	if ov.IsUnknown() {
		return customtypes.NewAuthtypesOidcconfigValueUnknown(), nil
	}
	return customtypes.NewAuthtypesOidcconfigValue(
		customtypes.AuthtypesOidcconfigValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
}

func castToSamlConfigTyped(ctx context.Context, ov basetypes.ObjectValue) (customtypes.AuthtypesSamlConfigValue, diag.Diagnostics) {
	if ov.IsNull() {
		return customtypes.NewAuthtypesSamlConfigValueNull(), nil
	}
	if ov.IsUnknown() {
		return customtypes.NewAuthtypesSamlConfigValueUnknown(), nil
	}
	return customtypes.NewAuthtypesSamlConfigValue(
		customtypes.AuthtypesSamlConfigValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
}

func castToRoleMappingTyped(ctx context.Context, ov basetypes.ObjectValue) (customtypes.AuthtypesRoleMappingValue, diag.Diagnostics) {
	if ov.IsNull() {
		return customtypes.NewAuthtypesRoleMappingValueNull(), nil
	}
	if ov.IsUnknown() {
		return customtypes.NewAuthtypesRoleMappingValueUnknown(), nil
	}
	return customtypes.NewAuthtypesRoleMappingValue(
		customtypes.AuthtypesRoleMappingValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
}

func castToAuthNProviderInfoTyped(ctx context.Context, ov basetypes.ObjectValue) (customtypes.AuthtypesAuthNproviderInfoValue, diag.Diagnostics) {
	if ov.IsNull() {
		return customtypes.NewAuthtypesAuthNproviderInfoValueNull(), nil
	}
	if ov.IsUnknown() {
		return customtypes.NewAuthtypesAuthNproviderInfoValueUnknown(), nil
	}
	return customtypes.NewAuthtypesAuthNproviderInfoValue(
		customtypes.AuthtypesAuthNproviderInfoValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
}

// buildAuthDomainConfigValue assembles the nested config customtype
// from already-flattened ObjectValue children.
func buildAuthDomainConfigValue(
	ctx context.Context,
	google, oidc, saml, role basetypes.ObjectValue,
	ssoEnabled types.Bool,
	ssoType types.String,
) (customtypes.AuthtypesAuthDomainConfigValue, diag.Diagnostics) {
	return customtypes.NewAuthtypesAuthDomainConfigValue(
		customtypes.AuthtypesAuthDomainConfigValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"google_auth_config": google,
			"oidc_config":        oidc,
			"role_mapping":       role,
			"saml_config":        saml,
			"sso_enabled":        ssoEnabled,
			"sso_type":           ssoType,
		},
	)
}
