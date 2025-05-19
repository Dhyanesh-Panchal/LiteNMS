package models

import (
	"errors"
	"github.com/lib/pq"
	"nms-backend/db"
)

var (
	ErrInvalidRequestBody = errors.New("invalid request body")
)

// DiscoveryProfile represents a discovery profile in the system
type DiscoveryProfile struct {
	ID int `json:"id" db:"discovery_profile_id"`

	DeviceIPs []string `json:"device_ips" db:"device_ips"`

	CredentialProfileIDs []int `json:"credential_profile_ids" db:"credential_profiles"`
}

func CreateDiscoveryProfile(db *db.ConfigDBClient, deviceIPs []string, credentialProfileIDs []int) (int, error) {
	query := `
		INSERT INTO discovery_profile (device_ips, credential_profiles)
		VALUES ($1, $2)
		RETURNING discovery_profile_id`

	var id int

	err := db.QueryRow(query, pq.Array(deviceIPs), pq.Array(credentialProfileIDs)).Scan(&id)

	return id, err
}

func GetAllDiscoveryProfiles(db *db.ConfigDBClient) ([]DiscoveryProfile, error) {
	query := `SELECT discovery_profile_id, device_ips, credential_profiles FROM discovery_profile`

	rows, err := db.Query(query)

	if err != nil {

		return nil, err

	}

	defer rows.Close()

	var profiles []DiscoveryProfile

	for rows.Next() {

		var profile DiscoveryProfile

		var credentialProfiles []int64

		if err = rows.Scan(&profile.ID, pq.Array(&profile.DeviceIPs), pq.Array(&credentialProfiles)); err != nil {

			return nil, err

		}

		profile.CredentialProfileIDs = make([]int, len(credentialProfiles))

		for i := range len(credentialProfiles) {

			profile.CredentialProfileIDs[i] = int(credentialProfiles[i])

		}

		profiles = append(profiles, profile)

	}

	return profiles, nil
}

func GetDiscoveryProfile(db *db.ConfigDBClient, id int) (DiscoveryProfile, error) {
	query := `SELECT discovery_profile_id, device_ips, credential_profiles FROM discovery_profile WHERE discovery_profile_id = $1`

	var profile DiscoveryProfile

	var credentialProfiles []int64

	err := db.QueryRow(query, id).Scan(&profile.ID, pq.Array(&profile.DeviceIPs), pq.Array(&credentialProfiles))

	if err != nil {

		return profile, err

	}

	profile.CredentialProfileIDs = make([]int, len(credentialProfiles))

	for i := range len(credentialProfiles) {

		profile.CredentialProfileIDs[i] = int(credentialProfiles[i])

	}

	return profile, nil
}

func UpdateDiscoveryProfile(db *db.ConfigDBClient, id int, deviceIPs []string, credentialProfileIDs []int) (int64, error) {
	query := `
		UPDATE discovery_profile
		SET device_ips = $1, credential_profiles = $2
		WHERE discovery_profile_id = $3`

	result, err := db.Exec(query, pq.Array(deviceIPs), pq.Array(credentialProfileIDs), id)

	if err != nil {

		return -1, err

	}

	return result.RowsAffected()
}
