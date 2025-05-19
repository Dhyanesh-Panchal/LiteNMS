package models

import (
	"github.com/lib/pq"
	"go.uber.org/zap"
	. "nms-backend/db"
	. "nms-backend/utils"
)

// CredentialProfile represents a credential profile in the system
type CredentialProfile struct {
	ID int `json:"id" db:"credential_profile_id"`

	Hostname string `json:"hostname" db:"hostname"`

	Password string `json:"password" db:"password"`

	Port uint16 `json:"port" db:"port"`
}

func CreateCredentialProfile(db *ConfigDBClient, hostname string, password string, port uint16) (int, error) {
	query := `
		INSERT INTO credential_profiles (hostname, password, port)
		VALUES ($1, $2, $3)
		RETURNING credential_profile_id`

	var id int

	err := db.QueryRow(query, hostname, password, port).Scan(&id)

	return id, err
}

func GetAllCredentialProfiles(db *ConfigDBClient) ([]CredentialProfile, error) {
	query := `SELECT credential_profile_id, hostname, password, port FROM credential_profiles`

	rows, err := db.Query(query)

	if err != nil {

		Logger.Error("Error querying credential profiles", zap.Error(err))

		return nil, err

	}

	defer rows.Close()

	var profiles []CredentialProfile

	for rows.Next() {

		var profile CredentialProfile

		if err = rows.Scan(&profile.ID, &profile.Hostname, &profile.Password, &profile.Port); err != nil {

			Logger.Error("Error scanning credential profile", zap.Error(err))

			return nil, err

		}

		profiles = append(profiles, profile)

	}

	return profiles, nil
}

func GetCredentialProfiles(db *ConfigDBClient, ids []int) ([]CredentialProfile, error) {
	query := `SELECT credential_profile_id, hostname, password, port FROM credential_profiles WHERE credential_profile_id = ANY($1)`

	var profiles []CredentialProfile

	rows, err := db.Query(query, pq.Array(ids))

	if err != nil {

		Logger.Error("Error querying credential profiles", zap.Error(err))

		return nil, err

	}

	defer rows.Close()

	for rows.Next() {

		var profile CredentialProfile

		if err = rows.Scan(&profile.ID, &profile.Hostname, &profile.Password, &profile.Port); err != nil {

			Logger.Error("Error scanning credential profile", zap.Error(err))

			return nil, err

		}

		profiles = append(profiles, profile)

	}

	return profiles, nil

}

func UpdateCredentialProfile(db *ConfigDBClient, id int, hostname string, password string, port uint16) (int64, error) {
	query := `
		UPDATE credential_profiles
		SET hostname = $1, password = $2, port = $3
		WHERE credential_profile_id = $4`

	result, err := db.Exec(query, hostname, password, port, id)

	if err != nil {

		Logger.Error("Error updating credential profile", zap.Error(err))

		return -1, err

	}

	return result.RowsAffected()
}
