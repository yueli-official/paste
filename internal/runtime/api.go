package runtime

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
	goframeapi "github.com/yueli-official/foundation/go/goframe/api"
	"github.com/yueli-official/foundation/go/goframe/ratelimit"
	"github.com/yueli-official/paste/internal/pasteerr"
)

const TraceHeader = "X-Trace-Id"

func MustAPIMiddleware() *goframeapi.Middleware {
	limit := 300
	if raw := os.Getenv("PASTE_RATE_LIMIT_PER_MINUTE"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			panic(fmt.Errorf("PASTE_RATE_LIMIT_PER_MINUTE must be a non-negative integer"))
		}
		limit = parsed
	}
	limiter := ratelimit.MustNew(ratelimit.Policy{Limit: limit, Window: time.Minute})
	middleware, err := goframeapi.New(goframeapi.Options{
		TraceHeader: TraceHeader,
		RateLimited: pasteerr.RateLimited,
		Validation:  pasteerr.Validation,
		Internal:    pasteerr.Internal,
		Limiter:     limiter,
		ClientKey:   func(request *ghttp.Request) string { return request.GetClientIp() },
	})
	if err != nil {
		panic(err)
	}
	return middleware
}

func TraceRoute(request *ghttp.Request) {
	if request.Header.Get(TraceHeader) == "" {
		request.Header.Set(TraceHeader, guid.S())
	}
	goframeapi.TraceRoute(request)
}
