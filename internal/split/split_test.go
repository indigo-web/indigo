package split

import (
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestSplit_MultipleSeparators(t *testing.T) {
	sample := "Hello World Yes?"
	iterator := StringIter(sample, ' ')
	result, err := iterator()
	require.NoError(t, err)
	require.Equal(t, "Hello", result)
	result, err = iterator()
	require.NoError(t, err)
	require.Equal(t, "World", result)
	result, err = iterator()
	require.NoError(t, err)
	require.Equal(t, "Yes?", result)
	result, err = iterator()
	require.EqualError(t, err, io.EOF.Error())
}

func TestSplit_NoSeparator(t *testing.T) {
	sample := "Hello,World!"
	iterator := StringIter(sample, ' ')
	result, err := iterator()
	require.NoError(t, err)
	require.Equal(t, "Hello,World!", result)
	result, err = iterator()
	require.EqualError(t, err, io.EOF.Error())
}

func TestSplit_SeparatorsOneByOne(t *testing.T) {
	sample := " Hello  World! "
	iterator := StringIter(sample, ' ')
	result, err := iterator()
	require.NoError(t, err)
	require.Equal(t, "", result)
	result, err = iterator()
	require.NoError(t, err)
	require.Equal(t, "Hello", result)
	result, err = iterator()
	require.NoError(t, err)
	require.Equal(t, "", result)
	result, err = iterator()
	require.NoError(t, err)
	require.Equal(t, "World!", result)
	result, err = iterator()
	require.NoError(t, err)
	require.Equal(t, "", result)
	result, err = iterator()
	require.EqualError(t, err, io.EOF.Error())
}
