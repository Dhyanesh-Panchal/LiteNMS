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

func Discover(discoveryIps []string, credentialProfiles []CredentialProfile) []Device {

	var discoveredDevices []Device

	lock := &sync.Mutex{}

	discoveryStatus := make(map[string]bool, len(discoveryIps))

	for _, credentialProfile := range credentialProfiles {

		config := &ssh.ClientConfig{

			User: credentialProfile.Hostname,

			Auth: []ssh.AuthMethod{

				ssh.Password(credentialProfile.Password),
			},

			HostKeyCallback: ssh.InsecureIgnoreHostKey(),

			Timeout: time.Second * 3,
		}

		var discoveryWorkersWaitGroup sync.WaitGroup

		for _, ip := range discoveryIps {

			if discovered := discoveryStatus[ip]; discovered {

				// device already discovered hence skip

				continue

			}

			discoveryWorkersWaitGroup.Add(1)

			go func() {

				defer discoveryWorkersWaitGroup.Done()

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

		discoveryWorkersWaitGroup.Wait()

	}

	Logger.Debug("discovery for", zap.Any("discoveredDevices", discoveredDevices))

	return discoveredDevices

}

func discoverDevice(ip string, credentialProfile CredentialProfile, config *ssh.ClientConfig) bool {

	client, err := ssh.Dial("tcp", ip+":"+strconv.Itoa(int(credentialProfile.Port)), config)

	if err != nil {

		Logger.Info("Failed to initialize the client", zap.Error(err))

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

	Logger.Info("Discovery Successful for", zap.String("ip", ip), zap.Any("Credential", credentialProfile), zap.String("Response", string(resp)))

	return true

}
