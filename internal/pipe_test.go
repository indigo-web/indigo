package internal

import (
	"bytes"
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

	pipe := NewPipe()

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

	pipe := NewPipe()

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

func TestPipeReadable(t *testing.T) {
	pipe := NewPipe()

	if pipe.Readable() {
		t.Fatal("empty pipe but readable, wanted to be false")
	}

	go func() {
		pipe.Write([]byte("Hello, world!"))
	}()

	// oh fuck, how dirty... But I really don't know how else to notify
	<-time.After(50 * time.Millisecond)

	if !pipe.Readable() {
		t.Fatal("pipe is not empty but not readable, wanted to be true")
	}

	_, _ = pipe.Read()

	if pipe.Readable() {
		t.Fatal("pipe is now again empty but readable, wanted to be false")
	}
}
