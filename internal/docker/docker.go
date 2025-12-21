package docker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Service struct {
	ctx context.Context
}

func NewService() *Service {
	return &Service{
		ctx: context.Background(),
	}
}

// VerifyContainer checks if a Docker container exists and is running
func (s *Service) VerifyContainer(containerName string) error {
	cmd := exec.CommandContext(s.ctx, "docker", "inspect", "--format={{.State.Running}}", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("container '%s' not found: %w", containerName, err)
	}

	isRunning := strings.TrimSpace(string(output))
	if isRunning != "true" {
		return fmt.Errorf("container '%s' is not running", containerName)
	}

	return nil
}

// Exec executes a command in the specified container
func (s *Service) Exec(containerName string, command []string) ([]byte, error) {
	args := append([]string{"exec", containerName}, command...)
	cmd := exec.CommandContext(s.ctx, "docker", args...)
	return cmd.CombinedOutput()
}
