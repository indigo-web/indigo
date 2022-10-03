package obtainer

import (
	"context"
	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/url"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
	settings2 "github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func nopHandler(context.Context, *types.Request) types.Response {
	return types.OK()
}

func newRequest(path string, method methods.Method) *types.Request {
	manager := headers.NewManager(settings2.Default().Headers)
	request, _ := types.NewRequest(
		&manager, url.Query{}, nil,
	)
	request.Path = path
	request.Method = method

	return request
}

func testPositiveMatch(t *testing.T, obtainer Obtainer) {
	path := "/"
	request := newRequest(path, methods.GET)
	_, methodsMap, err := obtainer(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, methodsMap)
}

func testNegativeMatchNotFound(t *testing.T, obtainer Obtainer) {
	path := "/42"
	request := newRequest(path, methods.GET)
	_, handler, err := obtainer(context.Background(), request)
	require.EqualError(t, err, http.ErrNotFound.Error())
	require.Nil(t, handler)
}

func testNegativeMatchMethodNotAllowed(t *testing.T, obtainer Obtainer) {
	path := "/"
	request := newRequest(path, methods.POST)
	ctx, handler, err := obtainer(context.Background(), request)
	require.EqualError(t, err, http.ErrMethodNotAllowed.Error())
	require.Nil(t, handler)

	allow := ctx.Value("allow")
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
