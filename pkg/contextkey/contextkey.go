package contextkey

type Key string

const (
	AccessTokenClaimsKey Key = "access_token_claims"
	UserIDKey            Key = "user_id"
	RolesKey             Key = "roles"

	JTIUsed    Key = "used"
	JTIRevoked Key = "revoked"

	AuthRefreshToken = "auth_refresh_token"

	// Casbin user action
	ActionRead   Key = "read"
	ActionUpdate Key = "update"
	ActionCreate Key = "create"
	ActionDelete Key = "delete"
)

const (
	HeaderDeviceConnectionToken = "X-Device-Connection-Token"
	HeaderDeviceRefreshToken    = "X-Device-Refresh-Token"
)

const (
	AudienceDeviceService = "device-service"
)
