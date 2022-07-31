package internal

import (
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestPipeReadWrite(t *testing.T) {
	wantedData := []byte("Hello!")
	pipe := NewPipe()

	go func() {
		pipe.Write(wantedData)
	}()

	data, err := pipe.Read()
	require.Nil(t, err)
	require.Equal(t, wantedData, data, "invalid data from pipe")
}

func TestPipeWriteErr(t *testing.T) {
	wantedErr := io.EOF
	pipe := NewPipe()

	go func() {
		pipe.WriteErr(wantedErr)
	}()

	_, err := pipe.Read()
	require.Error(t, err, wantedErr)
}

func TestPipeErrAfterWrite(t *testing.T) {
	wantedData := []byte("Hello!")
	wantedErr := io.EOF
	pipe := NewPipe()

	go func() {
		pipe.Write(wantedData)
		pipe.WriteErr(wantedErr)
	}()

	data, err := pipe.Read()
	require.Nil(t, err)
	require.Equal(t, wantedData, data, "invalid data from pipe")

	data, err = pipe.Read()
	require.Error(t, err, wantedErr)
}
