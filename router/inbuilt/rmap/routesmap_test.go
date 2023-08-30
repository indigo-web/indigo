package rmap

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRMap(t *testing.T) {
	t.Run("add single value", func(t *testing.T) {
		rmap := New()
		rmap.Add("/", method.GET, http.Respond)
		methodsMap, allow, ok := rmap.Get("/")
		require.True(t, ok)
		require.NotNil(t, methodsMap[method.GET])
		require.Equal(t, "GET", allow)
	})

	t.Run("add multiple methods to a single route", func(t *testing.T) {
		rmap := New()
		rmap.Add("/", method.GET, http.Respond)
		rmap.Add("/", method.POST, http.Respond)
		methodsMap, allow, ok := rmap.Get("/")
		require.True(t, ok)
		require.NotNil(t, methodsMap[method.GET])
		require.NotNil(t, methodsMap[method.POST])
		require.Equal(t, "GET,POST", allow)
	})

	t.Run("add multiple routes without growing", func(t *testing.T) {
		rmap := New()
		rmap.Add("/hello", method.GET, http.Respond)
		rmap.Add("/", method.GET, http.Respond)
		methodsMap, allow, ok := rmap.Get("/hello")
		require.True(t, ok)
		require.NotNil(t, methodsMap[method.GET])
		require.Equal(t, "GET", allow)
		methodsMap, allow, ok = rmap.Get("/")
		require.True(t, ok)
		require.NotNil(t, methodsMap[method.GET])
		require.Equal(t, "GET", allow)
	})
}
