package obtainer

import (
	"reflect"
	"testing"

	"github.com/indigo-web/indigo/internal/server/tcp/dummy"

	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/parser/http1"
	"github.com/indigo-web/indigo/settings"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"github.com/stretchr/testify/require"
)

func nopHandler(request *http.Request) http.Response {
	return http.RespondTo(request)
}

func newRequest(path string, method method.Method) *http.Request {
	hdrs := headers.NewHeaders(make(map[string][]string))
	bodyReader := http1.NewBodyReader(dummy.NewNopClient(), settings.Default().Body)
	request := http.NewRequest(
		hdrs, query.Query{}, http.NewResponse(), dummy.NewNopConn(), bodyReader, nil, false,
	)
	request.Path.String = path
	request.Method = method

	return request
}

func testPositiveMatch(t *testing.T, obtainer Obtainer) {
	path := "/"
	request := newRequest(path, method.GET)
	methodsMap, err := obtainer(request)
	require.NoError(t, err)
	require.NotNil(t, methodsMap)
}

func testNegativeMatchNotFound(t *testing.T, obtainer Obtainer) {
	path := "/42"
	request := newRequest(path, method.GET)
	handler, err := obtainer(request)
	require.EqualError(t, err, status.ErrNotFound.Error())
	require.Nil(t, handler)
}

func testNegativeMatchMethodNotAllowed(t *testing.T, obtainer Obtainer) {
	path := "/"
	request := newRequest(path, method.POST)
	handler, err := obtainer(request)
	require.EqualError(t, err, status.ErrMethodNotAllowed.Error())
	require.Nil(t, handler)

	allow := request.Ctx.Value("allow")
	require.NotNil(t, allow)
	require.Equal(t, "GET", allow.(string))
}

func TestStaticObtainer(t *testing.T) {
	routes := types.RoutesMap{
		"/": types.MethodsMap{
			method.GET: &types.HandlerObject{
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
	routes := types.RoutesMap{
		"/": types.MethodsMap{
			method.GET: &types.HandlerObject{
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
