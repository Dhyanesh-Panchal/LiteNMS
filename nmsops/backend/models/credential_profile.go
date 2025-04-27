package models

// CredentialProfile represents a credential profile in the system
type CredentialProfile struct {
	ID       int    `json:"id" db:"credential_profile_id"`
	Hostname string `json:"hostname" db:"hostname"`
	Password string `json:"password" db:"password"`
	Port     int16  `json:"port" db:"port"`
}

// CredentialProfileRequest represents the request body for creating and updating a credential profile
type CredentialProfileRequest struct {
	Hostname string `json:"hostname" binding:"required"`
	Password string `json:"password" binding:"required"`
	Port     int16  `json:"port" binding:"required"`
}

// CredentialProfileResponse represents the response for credential profile endpoints
type CredentialProfileResponse struct {
	Profiles []CredentialProfile `json:"profiles"`
}
