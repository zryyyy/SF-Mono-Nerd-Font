package docker

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/zryyyy/Nerd-Font-Patcher/internal/config"
)

func Check() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return errors.New("required command not found: docker")
	}
	return nil
}

func Run(patcher config.PatcherConfig, inputDir, outputDir string) error {
	absInput, err := filepath.Abs(inputDir)
	if err != nil {
		return err
	}
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return err
	}

	args := []string{
		"run",
		"--rm",
		"-v", absInput + ":/in",
		"-v", absOutput + ":/out",
		"-e", fmt.Sprintf("PN=%d", patcher.Parallel),
		patcher.Image,
	}
	args = append(args, patcher.Args...)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker patcher failed: %w", err)
	}
	return nil
}
