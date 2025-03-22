package spotify

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// checkAndCleanupPort checks if the port is in use and attempts to kill any processes using it
func checkAndCleanupPort(port int) error {
	// Get process ID using the port
	pid, err := getProcessUsingPort(port)
	if err != nil {
		return fmt.Errorf("failed to check port usage: %v", err)
	}

	// If no process is using the port, return
	if pid == 0 {
		return nil
	}

	log.Printf("Found process (PID: %d) using port %d. Attempting to terminate...", pid, port)

	// Kill the process
	if err := killProcess(pid); err != nil {
		return fmt.Errorf("failed to kill process %d: %v", pid, err)
	}

	log.Printf("Successfully terminated process using port %d", port)
	return nil
}

// getProcessUsingPort returns the PID of the process using the specified port
func getProcessUsingPort(port int) (int, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin", "linux":
		// For macOS and Linux, use lsof
		cmd = exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	case "windows":
		// For Windows, use netstat
		cmd = exec.Command("netstat", "-ano", "|", "findstr", fmt.Sprintf(":%d", port))
	default:
		return 0, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If the command returns an error, it likely means no process is using the port
		return 0, nil
	}

	// Parse the output to find the PID
	pid := parseProcessID(string(output), runtime.GOOS)
	return pid, nil
}

// killProcess terminates a process by its PID
func killProcess(pid int) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = exec.Command("kill", "-9", strconv.Itoa(pid))
	case "windows":
		cmd = exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Run()
}

// parseProcessID extracts the process ID from the command output
func parseProcessID(output, os string) int {
	switch os {
	case "darwin", "linux":
		// Example output format:
		// COMMAND  PID     USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
		// main    12345    user    8u  IPv6  0x123456789abcdef      0t0  TCP *:8081 (LISTEN)
		lines := strings.Split(output, "\n")
		if len(lines) < 2 {
			return 0
		}
		fields := strings.Fields(lines[1])
		if len(fields) < 3 {
			return 0
		}
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			return 0
		}
		return pid

	case "windows":
		// Example output format:
		// Proto  Local Address          Foreign Address        State           PID
		// TCP    0.0.0.0:8081          0.0.0.0:0              LISTENING       12345
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "LISTENING") {
				fields := strings.Fields(line)
				if len(fields) < 5 {
					continue
				}
				pid, err := strconv.Atoi(fields[4])
				if err != nil {
					continue
				}
				return pid
			}
		}
		return 0

	default:
		return 0
	}
}
