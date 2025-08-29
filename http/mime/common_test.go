package mime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComplies(t *testing.T) {
	for _, tc := range []string{"", JSON, JSON + ";", JSON + ";param"} {
		require.True(t, Complies(JSON, tc))
	}
}
