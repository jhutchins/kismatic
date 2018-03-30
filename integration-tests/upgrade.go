package integration_tests

import (
	"encoding/json"
	"os/exec"

	"github.com/apprenda/kismatic/pkg/install"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type infoOutput struct {
	EarliestVersion string
	LatestVersion   string
	Nodes           []struct {
		Version string
	}
}

type versionOutput struct {
	Version string
}

func assertClusterVersionIsCurrent() {
	By("Calling ./kismatic version to get Kismatic version")
	cmd := exec.Command("./kismatic", "version", "-o", "json")
	out, err := cmd.Output()
	FailIfError(err)
	ver := versionOutput{}
	err = json.Unmarshal(out, &ver)
	FailIfError(err)
	assertClusterVersion(ver.Version)
}

func assertClusterVersion(version string) {
	By("Calling ./kismatic info to get the cluster's version")
	planFileName := "kismatic-testing.yaml"
	fp := install.FilePlanner{File: planFileName}
	planFromFile, err := fp.Read()
	if err != nil {
		FailIfError(err, "Couldn't read from plan")
	}
	clusterName := planFromFile.Cluster.Name
	importCmd := exec.Command("./kismatic", "import", planFileName)
	if err := importCmd.Run(); err != nil {
		FailIfError(err, "Couldn't import plan")
	}
	cmd := exec.Command("./kismatic", "info", clusterName, "-o", "json")
	out, err := cmd.Output()
	FailIfError(err)
	info := infoOutput{}
	err = json.Unmarshal(out, &info)
	FailIfError(err)

	Expect(info.EarliestVersion).To(Equal(version))
	Expect(info.LatestVersion).To(Equal(version))
	for _, n := range info.Nodes {
		Expect(n.Version).To(Equal(version))
	}
}
