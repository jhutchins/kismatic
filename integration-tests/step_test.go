package integration_tests

import (
	"os"
	"os/exec"

	"github.com/apprenda/kismatic/pkg/install"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("install step commands", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
	})

	Describe("Running the api server play against an existing cluster", func() {
		ItOnAWS("should return successfully [slow]", func(aws infrastructureProvisioner) {
			WithMiniInfrastructure(Ubuntu1604LTS, aws, func(node NodeDeets, sshKey string) {
				err := installKismaticMini(node, sshKey)
				Expect(err).ToNot(HaveOccurred())
				planFile := "kismatic-testing.yaml"
				fp := install.FilePlanner{File: planFile}
				planFromFile, err := fp.Read()
				if err != nil {
					Expect(err).ToNot(HaveOccurred())
				}
				name := planFromFile.Cluster.Name
				importCmd := exec.Command("./kismatic", "import", planFile)
				if err := importCmd.Run(); err != nil {
					Expect(err).ToNot(HaveOccurred())
				}
				c := exec.Command("./kismatic", "install", "step", name, "_kube-apiserver.yaml")
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				err = c.Run()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
