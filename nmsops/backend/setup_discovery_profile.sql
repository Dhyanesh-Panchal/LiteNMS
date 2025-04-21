-- Drop the existing table if it exists
DROP TABLE IF EXISTS discovery_profile;

-- Create discovery_profile table
CREATE TABLE discovery_profile (
    discovery_profile_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    device_ips INTEGER[] NOT NULL,
    credential_profiles INTEGER[] NOT NULL
);

-- Insert some sample data
INSERT INTO discovery_profile (device_ips, credential_profiles)
VALUES 
    (ARRAY[2130706433, 3232235777], ARRAY[1, 2])
ON CONFLICT (discovery_profile_id) DO NOTHING; 