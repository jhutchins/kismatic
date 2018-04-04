package integration_tests

import (
	"fmt"
	"os/exec"

	"github.com/apprenda/kismatic/pkg/install"
)

const (
	defaultPlanFile = "kismatic-testing.yaml"
)

type ClusterPlan struct {
	Cluster struct {
		Name                       string
		DisablePackageInstallation string `yaml:"disable_package_installation"`
		Networking                 struct {
			Type             string
			PodCIDRBlock     string `yaml:"pod_cidr_block"`
			ServiceCIDRBlock string `yaml:"service_cidr_block"`
		}
		Certificates struct {
			Expiry          string
			LocationCity    string `yaml:"location_city"`
			LocationState   string `yaml:"location_state"`
			LocationCountry string `yaml:"location_country"`
		}
		SSH struct {
			User string
			Key  string `yaml:"ssh_key"`
			Port int    `yaml:"ssh_port"`
		}
	}
	Etcd struct {
		ExpectedCount int `yaml:"expected_count"`
		Nodes         []NodePlan
	}
	Master struct {
		ExpectedCount         int `yaml:"expected_count"`
		Nodes                 []NodePlan
		LoadBalancedFQDN      string `yaml:"load_balanced_fqdn"`
		LoadBalancedShortName string `yaml:"load_balanced_short_name"`
	}
	Worker struct {
		ExpectedCount int `yaml:"expected_count"`
		Nodes         []NodePlan
	}
	Ingress struct {
		ExpectedCount int `yaml:"expected_count"`
		Nodes         []NodePlan
	}
	Storage struct {
		ExpectedCount int `yaml:"expected_count"`
		Nodes         []NodePlan
	}
	NFS struct {
		Volumes []NFSVolume `yaml:"nfs_volume"`
	}
}

type NodePlan struct {
	host       string
	ip         string
	internalip string
}

func runImport(pfs ...string) (string, error) {
	var planFile string
	if len(pfs) == 0 {
		planFile = defaultPlanFile
	} else if len(pfs) == 1 {
		planFile = pfs[0]
	} else {
		return "", fmt.Errorf("runImport can only be passed either 0 or 1 arguments. 0 uses the default testing plan file")
	}
	fp := install.FilePlanner{File: planFile}
	planFromFile, err := fp.Read()
	if err != nil {
		return "", fmt.Errorf("error reading plan prior to import: %v", err)
	}
	clusterName := planFromFile.Cluster.Name
	importCmd := exec.Command("./kismatic", "import", planFile)
	if err := importCmd.Run(); err != nil {
		return "", fmt.Errorf("error running kismatic import: %v", err)
	}
	return clusterName, nil
}
