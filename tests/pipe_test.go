package tests

import (
	"bytes"
	"io"
	"testing"

	"indigo/internal"
)

func TestPipeReadWrite(t *testing.T) {
	wantedData := []byte("Hello!")

	pipe := internal.NewPipe()

	go func() {
		pipe.Write(wantedData)
	}()

	data, err := pipe.Read()

	if err != nil {
		t.Errorf("got error: %s", err.Error())
		return
	}

	if !bytes.Equal(data, wantedData) {
		t.Errorf("got invalid data from pipe (wanted %s, got %s)", string(wantedData), string(data))
	}
}

func TestPipeWriteErr(t *testing.T) {
	wantedErr := io.EOF

	pipe := internal.NewPipe()

	go func() {
		pipe.WriteErr(wantedErr)
	}()

	_, err := pipe.Read()

	if err != wantedErr {
		t.Fatalf("wanted: %s, got: %s", wantedErr.Error(), err.Error())
	}
}

func TestPipeErrAfterWrite(t *testing.T) {
	wantedData := []byte("Hello!")
	wantedErr := io.EOF

	pipe := internal.NewPipe()

	go func() {
		pipe.Write(wantedData)
		pipe.WriteErr(wantedErr)
	}()

	data, err := pipe.Read()

	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
		return
	}

	if !bytes.Equal(data, wantedData) {
		t.Fatalf("data wanted: %s, got: %s", string(wantedData), string(data))
		return
	}

	data, err = pipe.Read()

	if err != wantedErr {
		t.Fatalf("wanted error: %s, got: %s", wantedErr.Error(), err.Error())
	}
}
