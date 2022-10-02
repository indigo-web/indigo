package obtainer

import (
	"context"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
	"github.com/fakefloordiv/indigo/types"
)

type Obtainer func(context.Context, *types.Request) (context.Context, routertypes.HandlerFunc, error)
