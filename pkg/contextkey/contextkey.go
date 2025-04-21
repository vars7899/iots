package contextkey

type Key string

const (
	AccessTokenClaimsKey Key = "access_token_claims"
	UserIDKey            Key = "user_id"
	RolesKey             Key = "roles"

	JTIUsed    Key = "used"
	JTIRevoked Key = "revoked"

	Auth_refreshToken = "auth_refresh_token"
)
