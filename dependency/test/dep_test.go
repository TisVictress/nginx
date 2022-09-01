package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestEntrypoint(t *testing.T) {
	spec.Run(t, "Entrypoint", testEntrypoint, spec.Report(report.Terminal{}))
}

func testEntrypoint(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("Compile Nginx Dependency", func() {
		it("tests dependency", func() {
			tarballPath := os.Getenv("TARBALL_PATH")
			version := os.Getenv("VERSION")

			Expect("Dockerfile").To(BeARegularFile())

			_, err := exec.Command("docker", "build", "-t", "test", ".").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			_, err = exec.Command("docker", "run", "-v", fmt.Sprintf("%s:/tarball_path", filepath.Dir(tarballPath)), "test", version, "--tarballPath", tarballPath).CombinedOutput()
			// fmt.Println("************************\n", string(output))
			Expect(err).NotTo(HaveOccurred())
		})
	})
}
