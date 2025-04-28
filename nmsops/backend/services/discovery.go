package services

import (
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	. "nms-backend/models"
	. "nms-backend/utils"
	"strconv"
	"sync"
	"time"
)

func RunDiscovery(discoveryIps []uint32, credentialProfiles []CredentialProfile) []Device {

	var discoveredDevices []Device

	lock := &sync.Mutex{}

	discoveryStatus := make(map[uint32]bool, len(discoveryIps))

	for _, credentialProfile := range credentialProfiles {

		config := &ssh.ClientConfig{

			User: credentialProfile.Hostname,

			Auth: []ssh.AuthMethod{

				ssh.Password(credentialProfile.Password),
			},

			HostKeyCallback: ssh.InsecureIgnoreHostKey(),

			Timeout: time.Second * 10,
		}

		var wg sync.WaitGroup

		for _, ip := range discoveryIps {

			if discovered := discoveryStatus[ip]; discovered {

				// device already discovered hence skip

				continue

			}

			wg.Add(1)

			go func() {

				defer wg.Done()

				result := discoverDevice(ip, credentialProfile, config)

				if result {

					lock.Lock()

					discoveredDevices = append(discoveredDevices, Device{
						IP:            ip,
						CredentialID:  credentialProfile.ID,
						IsProvisioned: false,
					})

					lock.Unlock()

					discoveryStatus[ip] = true

				}

			}()

		}

		wg.Wait()

	}

	return discoveredDevices

}

func discoverDevice(ip uint32, credentialProfile CredentialProfile, config *ssh.ClientConfig) bool {

	client, err := ssh.Dial("tcp", ConvertNumericToIp(ip)+":"+strconv.Itoa(int(credentialProfile.Port)), config)

	if err != nil {

		Logger.Error("Failed to initialize the client", zap.Error(err))

		return false

	}

	defer client.Close()

	session, err := client.NewSession()

	if err != nil {

		Logger.Error("Failed to create session", zap.Error(err))

		return false

	}

	defer session.Close()

	resp, err := session.Output("whoami")

	if err != nil {

		Logger.Error("Failed to execute command", zap.Error(err))

		return false

	}

	Logger.Info("Discovery Successful for", zap.String("ip", ConvertNumericToIp(ip)), zap.Any("Credential", credentialProfile), zap.String("Response", string(resp)))

	return true

}
