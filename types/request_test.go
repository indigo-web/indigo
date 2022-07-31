package types

import (
	"github.com/stretchr/testify/require"
	"indigo/http"
	"indigo/internal"
	"io"
	"testing"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type bodySlicer struct {
	source       []byte
	offset, step int
}

func (s *bodySlicer) Next() []byte {
	s.offset += s.step
	return s.source[s.offset-s.step : min(s.offset, len(s.source))]
}

func getRequest() (Request, internal.Pipe) {
	return NewRequest(nil, make(http.Headers), nil, 10)
}

func feeder(pipe internal.Pipe, body []byte, n int) {
	for i := 0; i < len(body); i += n {
		pipe.Write(body[i:min(i+n, len(body))])
	}

	pipe.WriteErr(io.EOF)
}

func testGetBody(t *testing.T, body []byte, n int) {
	request, pipe := getRequest()
	slicer := bodySlicer{
		source: body,
		step:   n,
	}
	go feeder(pipe, body, n)

	require.NoError(t, request.GetBody(
		func(b []byte) error {
			require.Equal(t, slicer.Next(), b)
			return nil
		},
		func(err error) {
			require.NoError(t, err)
		}),
	)
}

func testGetFullBody(t *testing.T, body []byte, n int) {
	request, pipe := getRequest()
	go feeder(pipe, body, n)
	gotBody, err := request.GetFullBody()
	require.NoError(t, err)
	require.Equal(t, body, gotBody)
}

func TestRequest_GetBody(t *testing.T) {
	someLongBody := []byte("Hello, World! Lorem ipsum. Hope, you are good!")

	t.Run("SingleBodyPerOnce", func(t *testing.T) {
		testGetBody(t, someLongBody, len(someLongBody))
	})

	t.Run("BodyBy5Chars", func(t *testing.T) {
		testGetBody(t, someLongBody, 5)
	})

	t.Run("BodyBy1Char", func(t *testing.T) {
		testGetBody(t, someLongBody, 1)
	})
}

func TestRequest_GetFullBody(t *testing.T) {
	someLongBody := []byte("Hello, World! Lorem ipsum. Hope, you are good!")

	t.Run("WholeBodyPerOnce", func(t *testing.T) {
		testGetFullBody(t, someLongBody, len(someLongBody))
	})

	t.Run("BodyBy5Chars", func(t *testing.T) {
		testGetFullBody(t, someLongBody, 5)
	})

	t.Run("BodyBy1Char", func(t *testing.T) {
		testGetFullBody(t, someLongBody, 1)
	})
}

func TestRequest_Reset(t *testing.T) {
	// no requires or asserts because the only purpose of
	// these tests is just not to catch a deadlock

	t.Run("OnEmptyBody", func(t *testing.T) {
		request, pipe := getRequest()
		go func() {
			pipe.WriteErr(io.EOF)
		}()
		request.body.Reset()
	})

	t.Run("OnNotEmptyBody", func(t *testing.T) {
		request, pipe := getRequest()
		wantBody := []byte("Hello, world!")
		completionChan := make(chan bool)
		go func() {
			pipe.Write(wantBody)
			pipe.WriteErr(io.EOF)
			completionChan <- true
		}()
		request.body.Reset()
		<-completionChan
	})
}
