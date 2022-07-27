package internal

import (
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"
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

func TestPipeReadable(t *testing.T) {
	pipe := NewPipe()
	require.False(t, pipe.Readable(), "empty pipe but readable")

	go func() {
		pipe.Write([]byte("Hello, world!"))
	}()

	// oh fuck, how dirty... But I really don't know how else to notify
	<-time.After(50 * time.Millisecond)

	require.True(t, pipe.Readable(), "pipe is not empty but not readable")

	_, _ = pipe.Read()
	require.False(t, pipe.Readable(), "pipe is now again empty but readable")
}
