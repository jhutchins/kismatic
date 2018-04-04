package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/apprenda/kismatic/pkg/store"
)

const (
	defaultDBName        = "clusterStates.db"
	defaultPlanName      = "kismatic-cluster.yaml"
	defaultClusterName   = "kubernetes"
	defaultGeneratedName = "generated"
	defaultRunsName      = "runs"
	defaultTimeout       = 10 * time.Second
	clustersBucket       = "kismatic"
	assetsFolder         = "clusters"
	defaultInsecurePort  = "8080"
	defaultSecurePort    = "8443"
)

type planFileNotFoundErr struct {
	filename string
}

func (e planFileNotFoundErr) Error() string {
	return fmt.Sprintf("Plan file not found at %q. If you don't have a plan file, you may generate one with 'kismatic install plan'", e.filename)
}

// Returns a path to a plan file, generated dir, and runs dir according to the clusterName
func generateDirsFromName(clusterName string) (string, string, string) {
	return filepath.Join(assetsFolder, clusterName, defaultPlanName), filepath.Join(assetsFolder, clusterName, defaultGeneratedName), filepath.Join(assetsFolder, clusterName, defaultRunsName)
}

// CheckClusterExists does a simple check to see if the cluster folder+plan file exists in clusters
func CheckClusterExists(name string) (bool, error) {
	files, err := ioutil.ReadDir(assetsFolder)
	if err != nil {
		return false, err
	}
	for _, finfo := range files {
		if finfo.Name() == name {
			possiblePlans, err := ioutil.ReadDir(filepath.Join(assetsFolder, finfo.Name()))
			if err != nil {
				return false, err
			}
			for _, possiblePlan := range possiblePlans {
				if possiblePlan.Name() == defaultPlanName {
					return true, nil
				}
			}
		}
	}
	return false, fmt.Errorf("Cluster with name %s not found. If you have a plan file, but your cluster doesn't exist, please run kismatic import PLAN_FILE_PATH.", name)
}

// CheckPlaybookExists does a check to make sure the step exists
func CheckPlaybookExists(play string) (bool, error) {
	plays, err := ioutil.ReadDir("ansible/playbooks")
	if err != nil {
		return false, err
	}
	for _, finfo := range plays {
		if finfo.Name() == play {
			return true, nil
		}
	}
	return false, fmt.Errorf("playbook %s not found", play)
}

func CreateStoreIfNotExists(path string) (store.ClusterStore, *log.Logger) {
	parent, _ := filepath.Split(path)
	logger := log.New(os.Stdout, "[kismatic] ", log.LstdFlags|log.Lshortfile)
	if err := os.MkdirAll(parent, 0700); err != nil {
		logger.Fatalf("Error creating store directory structure: %v", err)
	}
	// Create the store
	s, err := store.New(path, 0600, logger)
	if err != nil {
		logger.Fatalf("Error creating store: %v", err)
	}
	err = s.CreateBucket(clustersBucket)
	if err != nil {
		logger.Fatalf("Error creating bucket in store: %v", err)
	}
	clusterStore := store.NewClusterStore(s, clustersBucket)
	return clusterStore, logger
}
