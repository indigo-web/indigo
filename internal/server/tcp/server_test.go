package tcp

import (
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:16161")
	require.NoError(t, err)

	server := NewServer(listener)
	stopCh := make(chan struct{})
	go func() {
		_ = server.Start(nil)
		stopCh <- struct{}{}
	}()
	require.NoError(t, server.Stop())
	<-stopCh
}
