package integration_tests

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/retry"
)

func verifyHeapster(master NodeDeets, sshKey string) error {
	// create volumes for alertmanager, prometheus-server and grafana
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
	cmd := exec.Command("./kismatic", "volume", "add", name, "1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error adding volume: %v", err)
	}

	// copy PVCs
	pvcs := []string{"influxdb-pvc.yaml"}
	for _, f := range pvcs {
		if err := copyFileToRemote(fmt.Sprintf("test-resources/heapster/%s", f), fmt.Sprintf("/tmp/%s", f), master, sshKey, 1*time.Minute); err != nil {
			return fmt.Errorf("error copying %s: %v", f, err)
		}
	}

	// create PVCs
	for _, f := range pvcs {
		if err := runViaSSH([]string{fmt.Sprintf("sudo kubectl --kubeconfig /root/.kube/config apply -f /tmp/%s", f)}, []NodeDeets{master}, sshKey, 1*time.Minute); err != nil {
			return fmt.Errorf("error creating pvc %s: %v", f, err)
		}
	}

	// verify pods are up
	deployments := map[string]int{
		"heapster":          3,
		"heapster-influxdb": 1,
	}
	return verifyDeployment(deployments, master, sshKey)
}

func verifyTiller(master NodeDeets, sshKey string) error {
	// verify pods are up
	deployments := map[string]int{
		"tiller-deploy": 1,
	}
	return verifyDeployment(deployments, master, sshKey)
}

func verifyDeployment(deployments map[string]int, master NodeDeets, sshKey string) error {
	for k, v := range deployments {
		if err := retry.WithBackoff(func() error {
			return runViaSSH([]string{fmt.Sprintf("sudo kubectl --kubeconfig /root/.kube/config get deployment %s -n kube-system -o jsonpath='{.status.availableReplicas}' | grep %d", k, v)}, []NodeDeets{master}, sshKey, 1*time.Minute)
		}, 10); err != nil {
			return fmt.Errorf("error verifying deployment %s: %v", k, err)
		}
	}

	return nil
}
