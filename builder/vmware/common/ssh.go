package common

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/packer/helper/multistep"
)

func CommHost(config *SSHConfig) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		driver := state.Get("driver").(Driver)

		if config.Comm.SSHHost != "" {
			return config.Comm.SSHHost, nil
		}

		ipAddrs, err := driver.PotentialGuestIP(state)
		if err != nil {
			log.Printf("IP lookup failed: %s", err)
			return "", fmt.Errorf("IP lookup failed: %s", err)
		}

		if len(ipAddrs) == 0 {
			log.Println("IP is blank, no IP yet.")
			return "", errors.New("IP is blank")
		}

		// Iterate through our list of addresses and dial each one. This way we
		// can dial up each one to see which lease is actually correct and has
		// ssh up.
		for index, ipAddress := range ipAddrs {
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ipAddress, config.Comm.SSHPort))

			// If we got a connection, then we should be good to go. Return the
			// address to the caller and pray that things work out.
			if err == nil {
				conn.Close()

				log.Printf("Detected IP: %s", ipAddress)
				return ipAddress, nil

			}

			// Otherwise we need to iterate to the next entry and keep hoping.
			log.Printf("Ignoring entry %d at %s:%d due to host being down.", index, ipAddress, config.Comm.SSHPort)
		}

		return "", errors.New("Host is not up")
	}
}
