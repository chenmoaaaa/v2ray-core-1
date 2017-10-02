// +build coverage

package scenarios

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/whatedcgveg/v2ray-core/common/uuid"
)

func BuildV2Ray() error {
	genTestBinaryPath()
	if _, err := os.Stat(testBinaryPath); err == nil {
		return nil
	}

	cmd := exec.Command("go", "test", "-tags", "json coverage coveragemain", "-coverpkg", "github.com/whatedcgveg/v2ray-core/...", "-c", "-o", testBinaryPath, GetSourcePath())
	return cmd.Run()
}

func RunV2RayProtobuf(config []byte) *exec.Cmd {
	genTestBinaryPath()

	covDir := filepath.Join(os.Getenv("GOPATH"), "out", "v2ray", "cov")
	os.MkdirAll(covDir, os.ModeDir)
	profile := uuid.New().String() + ".out"
	proc := exec.Command(testBinaryPath, "-config=stdin:", "-format=pb", "-test.run", "TestRunMainForCoverage", "-test.coverprofile", profile, "-test.outputdir", covDir)
	proc.Stdin = bytes.NewBuffer(config)
	proc.Stderr = os.Stderr
	proc.Stdout = os.Stdout

	return proc
}