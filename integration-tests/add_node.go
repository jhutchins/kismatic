package integration_tests

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/apprenda/kismatic/pkg/install"

	. "github.com/onsi/ginkgo"
)

func addNodeToCluster(newNode NodeDeets, sshKey string, labels []string, roles []string) error {
	By("Adding new worker")
	planFile := "kismatic-testing.yaml"
	fp := install.FilePlanner{File: planFile}
	plan, err := fp.Read()
	if err != nil {
		return fmt.Errorf("error reading plan: %v", err)
	}
	name := plan.Cluster.Name
	importCmd := exec.Command("./kismatic", "import", planFile)
	if err := importCmd.Run(); err != nil {
		return fmt.Errorf("error importing plan: %v", err)
	}
	cmd := exec.Command("./kismatic", "install", "add-node", name, "--roles", strings.Join(roles, ","), newNode.Hostname, newNode.PublicIP, newNode.PrivateIP)
	if len(labels) > 0 {
		cmd.Args = append(cmd.Args, "--labels", strings.Join(labels, ","))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running add node command: %v", err)
	}

	By("Verifying that the node was added")
	sshCmd := fmt.Sprintf("sudo kubectl --kubeconfig /root/.kube/config get nodes %s", strings.ToLower(newNode.Hostname)) // the api server is case-sensitive.
	out, err := executeCmd(sshCmd, newNode.PublicIP, newNode.SSHUser, sshKey)
	if err != nil {
		return fmt.Errorf("error getting nodes using kubectl: %v. Command output was: %s", err, out)
	}

	By("Verifying that the node is in the ready state")
	if !strings.Contains(strings.ToLower(out), "ready") {
		return fmt.Errorf("the node was not in ready state. node details: %s", out)
	}
	return nil
}
