package envdeployment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {

	require.True(t, true)

	t.Run("subtest example", func(t *testing.T) {
		require.True(t, true)
	})
}
