package obtainer

import (
	"context"
	"github.com/fakefloordiv/indigo/http/status"
	"reflect"
	"testing"

	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/url"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
	"github.com/stretchr/testify/require"
)

func nopHandler(request *http.Request) http.Response {
	return http.Respond(request)
}

func newRequest(path string, method methods.Method) *http.Request {
	hdrs := headers.NewHeaders(make(map[string][]string))
	request, _ := http.NewRequest(
		hdrs, url.Query{}, nil, context.Background(), http.NewResponse(),
	)
	request.Path = path
	request.Method = method

	return request
}

func testPositiveMatch(t *testing.T, obtainer Obtainer) {
	path := "/"
	request := newRequest(path, methods.GET)
	methodsMap, err := obtainer(request)
	require.NoError(t, err)
	require.NotNil(t, methodsMap)
}

func testNegativeMatchNotFound(t *testing.T, obtainer Obtainer) {
	path := "/42"
	request := newRequest(path, methods.GET)
	handler, err := obtainer(request)
	require.EqualError(t, err, status.ErrNotFound.Error())
	require.Nil(t, handler)
}

func testNegativeMatchMethodNotAllowed(t *testing.T, obtainer Obtainer) {
	path := "/"
	request := newRequest(path, methods.POST)
	handler, err := obtainer(request)
	require.EqualError(t, err, status.ErrMethodNotAllowed.Error())
	require.Nil(t, handler)

	allow := request.Ctx.Value("allow")
	require.NotNil(t, allow)
	require.Equal(t, "GET", allow.(string))
}

func TestStaticObtainer(t *testing.T) {
	routes := routertypes.RoutesMap{
		"/": routertypes.MethodsMap{
			methods.GET: &routertypes.HandlerObject{
				Fun: nopHandler,
			},
		},
	}
	obtainer := StaticObtainer(routes)

	t.Run("PositiveMatch", func(t *testing.T) {
		testPositiveMatch(t, obtainer)
	})

	t.Run("NegativeMatch_NotFound", func(t *testing.T) {
		testNegativeMatchNotFound(t, obtainer)
	})

	t.Run("NegativeMatch_MethodNotAllowed", func(t *testing.T) {
		testNegativeMatchMethodNotAllowed(t, obtainer)
	})
}

func TestDynamicObtainer(t *testing.T) {
	routes := routertypes.RoutesMap{
		"/": routertypes.MethodsMap{
			methods.GET: &routertypes.HandlerObject{
				Fun: nopHandler,
			},
		},
	}
	obtainer := DynamicObtainer(routes)

	t.Run("PositiveMatch", func(t *testing.T) {
		testPositiveMatch(t, obtainer)
	})

	t.Run("NegativeMatch_NotFound", func(t *testing.T) {
		testNegativeMatchNotFound(t, obtainer)
	})

	t.Run("NegativeMatch_MethodNotAllowed", func(t *testing.T) {
		testNegativeMatchMethodNotAllowed(t, obtainer)
	})
}

func TestAuto(t *testing.T) {
	onlyStatic := []string{
		"/hello/world",
		"/api/v1/",
		"/something/good",
	}

	someDynamic := []string{
		"/hello/world",
		"/still/static",
		"/finally/{dynamic part}",
	}

	staticObtainer := reflect.ValueOf(StaticObtainer).Pointer()
	dynamicObtainer := reflect.ValueOf(DynamicObtainer).Pointer()

	mustBeStatic := reflect.ValueOf(getObtainer(onlyStatic)).Pointer()
	mustBeDynamic := reflect.ValueOf(getObtainer(someDynamic)).Pointer()

	require.Equal(t, staticObtainer, mustBeStatic)
	require.Equal(t, dynamicObtainer, mustBeDynamic)
}
