package discovery

import (
	"encoding/json"
	"net/http"
	"slices"
	"time"
)

// https://openid.net/specs/openid-connect-discovery-1_0.html

const (
	SubjectTypePublic   = "public"
	SubjectTypePairwise = "pairwise"
)

const (
	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypeCIBA              = "urn:openid:params:grant-type:ciba"
	GrantTypeRefresh           = "refresh"
	GrantTypeImplicit          = "implicit"
)

const (
	TokenAuthClientSecretBasic = "client_secret_basic"
	TokenAuthClientSecretPost  = "client_secret_post"
	TokenAuthClientSecretJWT   = "client_secret_jwt"
	TokenAuthPrivateKeyJWT     = "private_key_jwt"
)

const (
	ResponseModeQuery    = "query"
	ResponseModeFragment = "fragment"
)

const (
	ResponseTypeCode    = "code"
	ResponseTypeIDToken = "id_token"
	ResponseTypeToken   = "token"
)

const (
	ACRUserVerification       = "urn:webauthn:verify"
	ACRUserPresence           = "urn:webauthn:presence"
	ACRPreferUserVerification = "urn:webauthn:prefer-verify"

	ACRPresenceInternal    = "urn:fido2:presence_internal"
	ACRFingerPrintInternal = "urn:fido2:fingerprint_internal"
	ACRPasscodeInternal    = "urn:fido2:passcode_internal"
	ACRVoiceprintInternal  = "urn:fido2:voiceprint_internal"
	ACRFaceprintInternal   = "urn:fido2:faceprint_internal"
	ACRLocationInternal    = "urn:fido2:location_internal"
	ACREyeprintInternal    = "urn:fido2:eyeprint_internal"
	ACRPatternInternal     = "urn:fido2:pattern_internal"
	ACRHandprintInternal   = "urn:fido2:handprint_internal"
	ACRPasscodeExternal    = "urn:fido2:passcode_external"
	ACRPatternExternal     = "urn:fido2:pattern_external"
)

const (
	AMRProofOfPossessionKey = "pop"
)

type Required struct {
	// URL using the https scheme with no query or fragment component that the OP asserts as its Issuer Identifier.
	Issuer string `json:"issuer,omitempty"`
	// URL of the OP's OAuth 2.0 Authorization Endpoint
	AuthorizationEndpoint string `json:"authorization_endpoint,omitempty"`
	// URL of the OP's OAuth 2.0 Token Endpoint
	TokenEndpoint string `json:"token_endpoint,omitempty"`

	// URL of the OP's JSON Web Key Set document.
	// This contains the signing key(s) the RP uses to validate signatures from the OP.
	// The JWK Set MAY also contain the Server's encryption key(s), which are used by RPs to encrypt requests to the Server.
	// When both signing and encryption keys are made available, a use (Key Use) parameter value is
	// REQUIRED for all keys in the referenced JWK Set to indicate each key's intended usage.
	// Although some algorithms allow the same key to be used for both signatures and encryption,
	// doing so is NOT RECOMMENDED, as it is less secure.
	// The JWK x5c parameter MAY be used to provide X.509 representations of keys provided.
	// When used, the bare key values MUST still be present and MUST match those in the certificate.
	JWKSURI string `json:"jwks_uri,omitempty"`

	// JSON array containing a list of the OAuth 2.0 scope values that this server supports.
	// The server MUST support the openid scope value.
	// Servers MAY choose not to advertise some supported scope values even when this parameter is used, although those defined in
	// OpenID.Core SHOULD be listed, if supported.
	ScopesSupported []string `json:"scopes_supported,omitempty"`

	// JSON array containing a list of the OAuth 2.0 response_type values that this OP supports.
	// Dynamic OpenID Providers MUST support the code, id_token, and the token id_token AssertionResponse Type values.
	ResponseTypesSupported []string `json:"response_types_supported,omitempty"`

	// JSON array containing a list of the Subject Identifier types that this OP supports.
	// Valid types include pairwise and public.
	SubjectTypesSupported []string `json:"subject_types_supported,omitempty"`

	// JSON array containing a list of the JWS signing algorithms (alg values) supported by the OP for the ID Token to encode the Claims in a JWT.
	// The algorithm RS256 MUST be included. The value none MAY be supported, but MUST NOT be used unless the AssertionResponse
	// Type used returns no ID Token from the Authorization Endpoint (such as when using the Authorization Code Flow).
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported,omitempty"`

	// JSON array containing a list of the OAuth 2.0 Grant Type values that this OP supports.
	// Dynamic OpenID Providers MUST support the authorization_code and implicit Grant Type values and MAY support other Grant Types.
	// If omitted, the default value is ["authorization_code", "implicit"].
	GrantTypesSupported []string `json:"grant_types_supported,omitempty"`

	// JSON array containing one or more of the following values: poll, ping, and push.
	BackChannelTokenDeliveryModesSupported []string `json:"backchannel_token_delivery_modes_supported,omitempty"`

	// URL of the OP's Backchannel Authentication Endpoint.
	BackChannelAuthenticationEndpoint string `json:"backchannel_authentication_endpoint,omitempty"`

	// URL of the OP's Backchannel QR Code Authentication Endpoint.
	BackChannelQRCodeAuthenticationEndpoint string `json:"backchannel_qr_code_authentication_endpoint,omitempty"`
}

type Optional struct {
	// URL of the OP's UserInfo Endpoint.
	UserInfoEndpoint string `json:"userinfo_endpoint,omitempty"`

	// URL of the OP's Dynamic Client Registration Endpoint
	RegistrationEndpoint string `json:"registration_endpoint,omitempty"`

	// JSON array containing a list of the OAuth 2.0 response_mode values that this OP supports, as specified in
	// OAuth 2.0 Multiple AssertionResponse Type Encoding Practices.
	// If omitted, the default for Dynamic OpenID Providers is ["query", "fragment"].
	ResponseModesSupported []string `json:"response_modes_supported,omitempty"`

	// JSON array containing a list of the Authentication Context Class References that this OP supports.
	ACRValuesSupported []string `json:"acr_values_supported,omitempty"`

	// JSON array containing a list of the JWE encryption algorithms (alg values) supported by the OP for the ID Token to encode the Claims in a JWT.
	IDTokenEncryptionAlgValuesSupported []string `json:"id_token_encryption_alg_values_supported,omitempty"`

	// JSON array containing a list of the JWE encryption algorithms (enc values) supported by the OP for the ID Token to encode the Claims in a JWT
	IDTokenEncryptionEncValuesSupported []string `json:"id_token_encryption_enc_values_supported,omitempty"`

	// JSON array containing a list of the JWS signing algorithms (alg values) supported by the UserInfo Endpoint to encode the Claims in a JWT.
	// The value none MAY be included.
	UserInfoSigningAlgValuesSupported []string `json:"userinfo_signing_alg_values_supported,omitempty"`

	// JSON array containing a list of the JWE encryption algorithms (alg values) supported by the UserInfo Endpoint to encode the Claims in a JWT.
	UserInfoEncryptionAlgValuesSupported []string `json:"userinfo_encryption_alg_values_supported,omitempty"`

	// JSON array containing a list of the JWE encryption algorithms (enc values) [JWA] supported by the UserInfo Endpoint to encode the Claims in a JWT
	UserInfoEncryptionEncValuesSupported []string `json:"userinfo_encryption_enc_values_supported,omitempty"`

	// JSON array containing a list of the JWS signing algorithms (alg values) supported by the OP for Request Objects,
	// which are described in Section 6.1 of OpenID Connect Core 1.0.
	// These algorithms are used both when the Request Object is passed by value (using the request parameter)
	// and when it is passed by reference (using the request_uri parameter). Servers SHOULD support none and RS256.
	RequestObjectSigningAlgValuesSupported []string `json:"request_object_signing_alg_values_supported,omitempty"`

	// SON array containing a list of the JWE encryption algorithms (alg values) supported by the OP for Request Objects.
	// These algorithms are used both when the Request Object is passed by value and when it is passed by reference.
	RequestObjectEncryptionAlgValuesSupported []string `json:"request_object_encryption_alg_values_supported,omitempty"`

	// JSON array containing a list of the JWE encryption algorithms (enc values) supported by the OP for Request Objects.
	// These algorithms are used both when the Request Object is passed by value and when it is passed by reference.
	RequestObjectEncryptionEncValuesSupported []string `json:"request_object_encryption_enc_values_supported,omitempty"`

	// JSON array containing a list of Client Authentication methods supported by this Token Endpoint.
	// The options are client_secret_post, client_secret_basic, client_secret_jwt, and private_key_jwt, as described in
	// Section 9 of OpenID Connect Core 1.0 [OpenID.Core].
	// Other authentication methods MAY be defined by extensions.
	// If omitted, the default is client_secret_basic -- the HTTP Basic Authentication Scheme specified in Section 2.3.1 of OAuth 2.0
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`

	// JSON array containing a list of the JWS signing algorithms (alg values) supported by the Token Endpoint for the signature on the JWT
	// used to authenticate the Client at the Token Endpoint for the private_key_jwt and client_secret_jwt authentication methods.
	// Servers SHOULD support RS256. The value none MUST NOT be used.
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty"`

	// JSON array containing a list of the display parameter values that the OpenID Provider supports.
	// These values are described in Section 3.1.2.1 of OpenID Connect Core 1.0
	DisplayValuesSupported []string `json:"display_values_supported,omitempty"`

	// JSON array containing a list of the Claim Types that the OpenID Provider supports.
	// These Claim Types are described in Section 5.6 of OpenID Connect Core 1.0.
	// Values defined by this specification are normal, aggregated, and distributed.
	// If omitted, the implementation supports only normal Claims.
	ClaimTypesSupported []string `json:"claim_types_supported,omitempty"`

	// JSON array containing a list of the Claim Names of the Claims that the OpenID Provider MAY be able to supply values for.
	// Note that for privacy or other reasons, this might not be an exhaustive list.
	ClaimsSupported []string `json:"claims_supported,omitempty"`

	// URL of a page containing human-readable information that developers might want or need to know when using the OpenID Provider.
	// In particular, if the OpenID Provider does not support Dynamic Client Registration,
	// then information on how to register Clients needs to be provided in this documentation.
	ServiceDocumentation string `json:"service_documentation,omitempty"`

	// Languages and scripts supported for values in Claims being returned, represented as a JSON array of BCP47 [RFC5646] language tag values.
	// Not all languages and scripts are necessarily supported for all Claim values.
	ClaimsLocalesSupported []string `json:"claims_locales_supported,omitempty"`

	// Languages and scripts supported for the user interface, represented as a JSON array of BCP47 [RFC5646] language tag values.
	UILocalesSupported []string `json:"ui_locales_supported,omitempty"`

	// Boolean value specifying whether the OP supports use of the claims parameter, with true indicating support.
	// If omitted, the default value is false.
	ClaimsParameterSupported bool `json:"claims_parameter_supported"`

	// Boolean value specifying whether the OP supports use of the request parameter, with true indicating support.
	// If omitted, the default value is false.
	RequestParameterSupported bool `json:"request_parameter_supported"`

	// Boolean value specifying whether the OP supports use of the request_uri parameter, with true indicating support.
	// If omitted, the default value is true.
	RequestURIParameterSupported bool `json:"request_uri_parameter_supported"`

	// Boolean value specifying whether the OP requires any request_uri values used to be pre-registered using the request_uris registration parameter.
	// Pre-registration is REQUIRED when the value is true. If omitted, the default value is false.
	RequireRequestURIRegistration bool `json:"require_request_uri_registration"`

	// URL that the OpenID Provider provides to the person registering the Client to read about the OP's requirements
	// on how the Relying Party can use the data provided by the OP.
	// The registration process SHOULD display this URL to the person registering the Client if it is given.
	OpPolicyURI string `json:"op_policy_uri,omitempty"`

	// URL that the OpenID Provider provides to the person registering the Client to read about OpenID Provider's terms of service.
	// The registration process SHOULD display this URL to the person registering the Client if it is given.
	OpTosURI string `json:"op_tos_uri,omitempty"`
}

type Full struct {
	// URL using the https scheme with no query or fragment component that the OP asserts as its Issuer Identifier.
	Issuer string `json:"issuer"`
	// URL of the OP's OAuth 2.0 Authorization Endpoint
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	// URL of the OP's OAuth 2.0 Token Endpoint
	TokenEndpoint string `json:"token_endpoint"`

	// URL of the OP's JSON Web Key Set document.
	// This contains the signing key(s) the RP uses to validate signatures from the OP.
	// The JWK Set MAY also contain the Server's encryption key(s), which are used by RPs to encrypt requests to the Server.
	// When both signing and encryption keys are made available, a use (Key Use) parameter value is
	// REQUIRED for all keys in the referenced JWK Set to indicate each key's intended usage.
	// Although some algorithms allow the same key to be used for both signatures and encryption,
	// doing so is NOT RECOMMENDED, as it is less secure.
	// The JWK x5c parameter MAY be used to provide X.509 representations of keys provided.
	// When used, the bare key values MUST still be present and MUST match those in the certificate.
	JWKSURI string `json:"jwks_uri"`

	// JSON array containing a list of the OAuth 2.0 scope values that this server supports.
	// The server MUST support the openid scope value.
	// Servers MAY choose not to advertise some supported scope values even when this parameter is used, although those defined in
	// OpenID.Core SHOULD be listed, if supported.
	ScopesSupported []string `json:"scopes_supported"`

	// JSON array containing a list of the OAuth 2.0 response_type values that this OP supports.
	// Dynamic OpenID Providers MUST support the code, id_token, and the token id_token AssertionResponse Type values.
	ResponseTypesSupported []string `json:"response_types_supported"`

	// JSON array containing a list of the Subject Identifier types that this OP supports.
	// Valid types include pairwise and public.
	SubjectTypesSupported []string `json:"subject_types_supported"`

	// JSON array containing a list of the JWS signing algorithms (alg values) supported by the OP for the ID Token to encode the Claims in a JWT.
	// The algorithm RS256 MUST be included. The value none MAY be supported, but MUST NOT be used unless the AssertionResponse
	// Type used returns no ID Token from the Authorization Endpoint (such as when using the Authorization Code Flow).
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`

	// URL of the OP's UserInfo Endpoint.
	UserInfoEndpoint string `json:"userinfo_endpoint,omitempty"`

	// URL of the OP's Dynamic Client Registration Endpoint
	RegistrationEndpoint string `json:"registration_endpoint,omitempty"`

	// JSON array containing a list of the OAuth 2.0 response_mode values that this OP supports, as specified in
	// OAuth 2.0 Multiple AssertionResponse Type Encoding Practices.
	// If omitted, the default for Dynamic OpenID Providers is ["query", "fragment"].
	ResponseModesSupported []string `json:"response_modes_supported,omitempty"`

	// JSON array containing a list of the OAuth 2.0 Grant Type values that this OP supports.
	// Dynamic OpenID Providers MUST support the authorization_code and implicit Grant Type values and MAY support other Grant Types.
	// If omitted, the default value is ["authorization_code", "implicit"].
	GrantTypesSupported []string `json:"grant_types_supported,omitempty"`

	// JSON array containing a list of the Authentication Context Class References that this OP supports.
	ACRValuesSupported []string `json:"acr_values_supported,omitempty"`

	// JSON array containing a list of the JWE encryption algorithms (alg values) supported by the OP for the ID Token to encode the Claims in a JWT.
	IDTokenEncryptionAlgValuesSupported []string `json:"id_token_encryption_alg_values_supported,omitempty"`

	// JSON array containing a list of the JWE encryption algorithms (enc values) supported by the OP for the ID Token to encode the Claims in a JWT
	IDTokenEncryptionEncValuesSupported []string `json:"id_token_encryption_enc_values_supported,omitempty"`

	// JSON array containing a list of the JWS signing algorithms (alg values) supported by the UserInfo Endpoint to encode the Claims in a JWT.
	// The value none MAY be included.
	UserInfoSigningAlgValuesSupported []string `json:"userinfo_signing_alg_values_supported,omitempty"`

	// JSON array containing a list of the JWE encryption algorithms (alg values) supported by the UserInfo Endpoint to encode the Claims in a JWT.
	UserInfoEncryptionAlgValuesSupported []string `json:"userinfo_encryption_alg_values_supported,omitempty"`

	// JSON array containing a list of the JWE encryption algorithms (enc values) [JWA] supported by the UserInfo Endpoint to encode the Claims in a JWT
	UserInfoEncryptionEncValuesSupported []string `json:"userinfo_encryption_enc_values_supported,omitempty"`

	// JSON array containing a list of the JWS signing algorithms (alg values) supported by the OP for Request Objects,
	// which are described in Section 6.1 of OpenID Connect Core 1.0.
	// These algorithms are used both when the Request Object is passed by value (using the request parameter)
	// and when it is passed by reference (using the request_uri parameter). Servers SHOULD support none and RS256.
	RequestObjectSigningAlgValuesSupported []string `json:"request_object_signing_alg_values_supported,omitempty"`

	// SON array containing a list of the JWE encryption algorithms (alg values) supported by the OP for Request Objects.
	// These algorithms are used both when the Request Object is passed by value and when it is passed by reference.
	RequestObjectEncryptionAlgValuesSupported []string `json:"request_object_encryption_alg_values_supported,omitempty"`

	// JSON array containing a list of the JWE encryption algorithms (enc values) supported by the OP for Request Objects.
	// These algorithms are used both when the Request Object is passed by value and when it is passed by reference.
	RequestObjectEncryptionEncValuesSupported []string `json:"request_object_encryption_enc_values_supported,omitempty"`

	// JSON array containing a list of Client Authentication methods supported by this Token Endpoint.
	// The options are client_secret_post, client_secret_basic, client_secret_jwt, and private_key_jwt, as described in
	// Section 9 of OpenID Connect Core 1.0 [OpenID.Core].
	// Other authentication methods MAY be defined by extensions.
	// If omitted, the default is client_secret_basic -- the HTTP Basic Authentication Scheme specified in Section 2.3.1 of OAuth 2.0
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`

	// JSON array containing a list of the JWS signing algorithms (alg values) supported by the Token Endpoint for the signature on the JWT
	// used to authenticate the Client at the Token Endpoint for the private_key_jwt and client_secret_jwt authentication methods.
	// Servers SHOULD support RS256. The value none MUST NOT be used.
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty"`

	// JSON array containing a list of the display parameter values that the OpenID Provider supports.
	// These values are described in Section 3.1.2.1 of OpenID Connect Core 1.0
	DisplayValuesSupported []string `json:"display_values_supported,omitempty"`

	// JSON array containing a list of the Claim Types that the OpenID Provider supports.
	// These Claim Types are described in Section 5.6 of OpenID Connect Core 1.0.
	// Values defined by this specification are normal, aggregated, and distributed.
	// If omitted, the implementation supports only normal Claims.
	ClaimTypesSupported []string `json:"claim_types_supported,omitempty"`

	// JSON array containing a list of the Claim Names of the Claims that the OpenID Provider MAY be able to supply values for.
	// Note that for privacy or other reasons, this might not be an exhaustive list.
	ClaimsSupported []string `json:"claims_supported,omitempty"`

	// URL of a page containing human-readable information that developers might want or need to know when using the OpenID Provider.
	// In particular, if the OpenID Provider does not support Dynamic Client Registration,
	// then information on how to register Clients needs to be provided in this documentation.
	ServiceDocumentation string `json:"service_documentation,omitempty"`

	// Languages and scripts supported for values in Claims being returned, represented as a JSON array of BCP47 [RFC5646] language tag values.
	// Not all languages and scripts are necessarily supported for all Claim values.
	ClaimsLocalesSupported []string `json:"claims_locales_supported,omitempty"`

	// Languages and scripts supported for the user interface, represented as a JSON array of BCP47 [RFC5646] language tag values.
	UILocalesSupported []string `json:"ui_locales_supported,omitempty"`

	// Boolean value specifying whether the OP supports use of the claims parameter, with true indicating support.
	// If omitted, the default value is false.
	ClaimsParameterSupported bool `json:"claims_parameter_supported"`

	// Boolean value specifying whether the OP supports use of the request parameter, with true indicating support.
	// If omitted, the default value is false.
	RequestParameterSupported bool `json:"request_parameter_supported"`

	// Boolean value specifying whether the OP supports use of the request_uri parameter, with true indicating support.
	// If omitted, the default value is true.
	RequestURIParameterSupported bool `json:"request_uri_parameter_supported"`

	// Boolean value specifying whether the OP requires any request_uri values used to be pre-registered using the request_uris registration parameter.
	// Pre-registration is REQUIRED when the value is true. If omitted, the default value is false.
	RequireRequestURIRegistration bool `json:"require_request_uri_registration"`

	// URL that the OpenID Provider provides to the person registering the Client to read about the OP's requirements
	// on how the Relying Party can use the data provided by the OP.
	// The registration process SHOULD display this URL to the person registering the Client if it is given.
	OpPolicyURI string `json:"op_policy_uri,omitempty"`

	// URL that the OpenID Provider provides to the person registering the Client to read about OpenID Provider's terms of service.
	// The registration process SHOULD display this URL to the person registering the Client if it is given.
	OpTosURI string `json:"op_tos_uri,omitempty"`

	// JSON array containing one or more of the following values: poll, ping, and push.
	BackChannelTokenDeliveryModesSupported []string `json:"backchannel_token_delivery_modes_supported,omitempty"`

	// URL of the OP's Backchannel Authentication Endpoint.
	BackChannelAuthenticationEndpoint string `json:"backchannel_authentication_endpoint,omitempty"`
}

func (f *Full) AddSupportedIDTokenSigningAlg(alg string) {
	if !slices.Contains(f.IDTokenSigningAlgValuesSupported, alg) {
		f.IDTokenSigningAlgValuesSupported = append(f.IDTokenSigningAlgValuesSupported, alg)
	}
}

func NewConfig(config *Required, optional *Optional) *Full {
	res := &Full{
		Issuer:                                     "",
		AuthorizationEndpoint:                      "",
		TokenEndpoint:                              "",
		JWKSURI:                                    "",
		ScopesSupported:                            []string{"openid"},
		ResponseTypesSupported:                     []string{"code"},
		SubjectTypesSupported:                      []string{SubjectTypePublic},
		IDTokenSigningAlgValuesSupported:           []string{"RS256"},
		UserInfoEndpoint:                           "",
		RegistrationEndpoint:                       "",
		ResponseModesSupported:                     []string{ResponseModeQuery, ResponseModeFragment},
		GrantTypesSupported:                        []string{GrantTypeAuthorizationCode},
		ACRValuesSupported:                         []string{},
		IDTokenEncryptionAlgValuesSupported:        []string{},
		IDTokenEncryptionEncValuesSupported:        []string{},
		UserInfoSigningAlgValuesSupported:          []string{},
		UserInfoEncryptionAlgValuesSupported:       []string{},
		UserInfoEncryptionEncValuesSupported:       []string{},
		RequestObjectSigningAlgValuesSupported:     []string{},
		RequestObjectEncryptionAlgValuesSupported:  []string{},
		RequestObjectEncryptionEncValuesSupported:  []string{},
		TokenEndpointAuthMethodsSupported:          []string{TokenAuthClientSecretBasic},
		TokenEndpointAuthSigningAlgValuesSupported: []string{},
		DisplayValuesSupported:                     []string{},
		ClaimTypesSupported:                        []string{},
		ClaimsSupported:                            []string{},
		ServiceDocumentation:                       "",
		ClaimsLocalesSupported:                     []string{},
		UILocalesSupported:                         []string{},
		ClaimsParameterSupported:                   false,
		RequestParameterSupported:                  false,
		RequestURIParameterSupported:               true,
		RequireRequestURIRegistration:              false,
		OpPolicyURI:                                "",
		OpTosURI:                                   "",
		BackChannelTokenDeliveryModesSupported:     nil,
		BackChannelAuthenticationEndpoint:          "",
	}
	x, _ := json.Marshal(config)
	_ = json.Unmarshal(x, res)
	if optional != nil {
		x, _ = json.Marshal(optional)
		_ = json.Unmarshal(x, &res)
	}
	return res
}

func Fetch(url string) (*Full, error) {
	cli := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout:   time.Second * 10,
			IdleConnTimeout:       time.Second * 10,
			ResponseHeaderTimeout: time.Second * 10,
			ExpectContinueTimeout: time.Second * 10,
		},
		Timeout: time.Second * 10,
	}
	resp, err := cli.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	res := &Full{}
	err = dec.Decode(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func FetchIssuer(issuer string) (*Full, error) {
	return Fetch(issuer + "/.well-known/openid-configuration")
}
