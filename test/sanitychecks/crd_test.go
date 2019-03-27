package sanitychecks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

type pkgYAML struct {
	Channels []struct {
		CurrentCSV string `yaml:"currentCSV"`
		Name       string `yaml:"name"`
	} `yaml:"channels"`
	PackageName string `yaml:"packageName"`
}

func TestLatestCRDFiles(t *testing.T) {
	curdir, err := os.Getwd()
	if err != nil {
		t.Error("Cannot locate test directory")
	}
	rootdir := filepath.Dir(filepath.Dir(curdir))
	crds, err := filepath.Glob(fmt.Sprintf("%s/deploy/crds/*_crd.yaml", rootdir))
	if err != nil {
		t.Error("Cannot locate CRD files inside: ", rootdir+"/deploy/crds")
	}

	content, err := ioutil.ReadFile(filepath.Join(rootdir, "manifests", "devopsconsole", "devopsconsole.package.yaml"))
	if err != nil {
		t.Error("Cannot read the devopsconsole.package.yaml")
	}

	pkg := &pkgYAML{}

	err = yaml.Unmarshal(content, pkg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	latestVersion := strings.Split(pkg.Channels[0].CurrentCSV, "operator.v")[1]

	for _, path := range crds {
		deployCRD, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		filename := filepath.Base(path)
		manifestPath := filepath.Join(rootdir, "manifests", "devopsconsole", latestVersion, filename)
		manifestCRD, err := ioutil.ReadFile(manifestPath)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		if bytes.Equal(deployCRD, manifestCRD) == false {
			t.Error("Files not matching: ", path, manifestPath)
		}
	}
}
