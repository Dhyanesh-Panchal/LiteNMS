package models

// DiscoveryProfile represents a discovery profile in the system
type DiscoveryProfile struct {
	ID                   int      `json:"id" db:"discovery_profile_id"`
	DeviceIPs            []uint32 `json:"device_ips" db:"device_ips"`
	CredentialProfileIDs []int    `json:"credential_profile_ids" db:"credential_profiles"`
}

// CreateDiscoveryProfileRequest represents the request body for creating a discovery profile
type CreateDiscoveryProfileRequest struct {
	DeviceIPs            []uint32 `json:"device_ips" binding:"required"`
	CredentialProfileIDs []int    `json:"credential_profile_ids" binding:"required"`
}

// UpdateDiscoveryProfileRequest represents the request body for updating a discovery profile
type UpdateDiscoveryProfileRequest struct {
	DeviceIPs            []uint32 `json:"device_ips" binding:"required"`
	CredentialProfileIDs []int    `json:"credential_profile_ids" binding:"required"`
}

// DiscoveryProfileResponse represents the response for discovery profile endpoints
type DiscoveryProfileResponse struct {
	Profiles []DiscoveryProfile `json:"profiles"`
}
