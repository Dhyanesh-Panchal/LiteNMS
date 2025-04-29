package models

// Device represents a network device in the system
type Device struct {
	IP            string `json:"ip" db:"ip"`
	CredentialID  int    `json:"credential_id" db:"credential_id"`
	IsProvisioned bool   `json:"is_provisioned" db:"is_provisioned"`
}

// DeviceResponse represents the response format for device API endpoints
type DeviceResponse struct {
	Devices []Device `json:"devices"`
}

type DeviceProvisionUpdateRequest struct {
	ProvisionUpdateIps []string `json:"provision_update_ips"`
}
