package tests

import (
	"bytes"
	"testing"

	"indigo/internal"
)

func readPipe(pipe *internal.Pipe, reschan chan []byte, errchan chan error) {
	data, err := pipe.Read()

	if err != nil {
		errchan <- err
		return
	}

	reschan <- data
}

func TestPipe(t *testing.T) {
	wantedData := []byte("Hello!")

	pipe := internal.NewPipe()
	reschan, errchan := make(chan []byte), make(chan error)
	go readPipe(pipe, reschan, errchan)
	go func() {
		pipe.Write(wantedData)
	}()

	select {
	case gotData := <-reschan:
		if !bytes.Equal(gotData, wantedData) {
			t.Errorf("got invalid data from pipe (wanted %s, got %s)", string(wantedData), string(gotData))
		}
		return
	case err := <-errchan:
		t.Errorf("got error: %s", err.Error())
		return
	}
}
