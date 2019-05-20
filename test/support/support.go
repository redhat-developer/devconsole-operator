package support

import (
	"os"
	"testing"
)

// Getenv returns a value of environment variable, if it exists.
// Returns the default value otherwise.
func Getenv(t *testing.T, key string, defaultValue string) string {
	value, found := os.LookupEnv(key)
	var retVal string
	if found {
		retVal = value
	} else {
		retVal = defaultValue
	}
	t.Logf("Using env variable: %s=%s", key, retVal)
	return retVal
}
