package radix

import (
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestArrMap(t *testing.T) {
	t.Run("no resize", func(t *testing.T) {
		var arrmap arrMap

		for i := 0; i < loadfactor; i++ {
			arrmap.Add(strconv.Itoa(i), new(Node))
		}

		require.False(t, arrmap.arrOverflow, "must not escape to map")
		require.Equal(t, loadfactor, len(arrmap.arr))
	})

	t.Run("with resize", func(t *testing.T) {
		var arrmap arrMap

		for i := 0; i < loadfactor+1; i++ {
			arrmap.Add(strconv.Itoa(i), new(Node))
		}

		require.True(t, arrmap.arrOverflow, "must escape to map")
		require.Equal(t, loadfactor+1, len(arrmap.m))
	})
}
