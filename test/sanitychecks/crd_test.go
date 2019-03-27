package sanitychecks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	rootdir := filepath.Dir(filepath.Dir(curdir))
	crds, err := filepath.Glob(fmt.Sprintf("%s/deploy/crds/*_crd.yaml", rootdir))
	require.NoErrorf(t, err, "Cannot locate CRD files inside: %s", rootdir+"/deploy/crds")

	content, err := ioutil.ReadFile(filepath.Join(rootdir, "manifests", "devopsconsole", "devopsconsole.package.yaml"))
	require.NoErrorf(t, err, "Cannot read the devopsconsole.package.yaml")

	pkg := &pkgYAML{}

	err = yaml.Unmarshal(content, pkg)
	require.NoError(t, err)

	latestVersion := strings.Split(pkg.Channels[0].CurrentCSV, "operator.v")[1]

	for _, path := range crds {
		deployCRD, err := ioutil.ReadFile(path)
		require.NoError(t, err)
		filename := filepath.Base(path)
		manifestPath := filepath.Join(rootdir, "manifests", "devopsconsole", latestVersion, filename)
		manifestCRD, err := ioutil.ReadFile(manifestPath)
		require.NoError(t, err)
		if bytes.Equal(deployCRD, manifestCRD) == false {
			t.Error("Files not matching: ", path, manifestPath)
		}
	}
}
