package models

// CredentialProfile represents a credential profile in the system
type CredentialProfile struct {
	ID       int    `json:"id" db:"credential_profile_id"`
	Hostname string `json:"hostname" db:"hostname"`
	Password string `json:"password" db:"password"`
	Port     int16  `json:"port" db:"port"`
}

// CreateCredentialProfileRequest represents the request body for creating a credential profile
type CreateCredentialProfileRequest struct {
	Hostname string `json:"hostname" binding:"required"`
	Password string `json:"password" binding:"required"`
	Port     int16  `json:"port" binding:"required"`
}

// UpdateCredentialProfileRequest represents the request body for updating a credential profile
type UpdateCredentialProfileRequest struct {
	Hostname string `json:"hostname" binding:"required"`
	Password string `json:"password" binding:"required"`
	Port     int16  `json:"port" binding:"required"`
}

// CredentialProfileResponse represents the response for credential profile endpoints
type CredentialProfileResponse struct {
	Profiles []CredentialProfile `json:"profiles"`
} 