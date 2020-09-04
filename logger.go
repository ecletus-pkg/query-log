package querylog

import (
	"fmt"
	"time"

	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"
)

func logger(out *Output, r *Rule, site *core.Site) func(action string, scope *aorm.Scope) {
	return func(action string, scope *aorm.Scope) {
		ctx := core.ContextFromDB(scope.DB())
		if !r.Accept(ctx, site, action, scope) {
			return
		}

		now := scope.ExecTime
		if now.IsZero() {
			now = time.Now()
		}

		var key = r.Key
		if key != "" {
			key = "[" + key + "] "
		}

		if ctx != nil {
			req := ctx.Request
			_, _ = fmt.Fprintf(out, "%v %s %v[%v %v] %v -> %v\n", now, site.Name(), key, req.URL, req.Method, action, scope.Query)
		} else {
			_, _ = fmt.Fprintf(out, "%v %s %v[%T] %v -> %v\n", now, site.Name(), key, scope.Value, action, scope.Query)
		}
		return
	}
}

func loggerError(out *Output, r *Rule, site *core.Site) func(scope *aorm.Scope, err error) {
	return func(scope *aorm.Scope, err error) {
		now := time.Now()
		ctx := core.ContextFromDB(scope.DB())

		var key = r.Key
		if key != "" {
			key = "[" + key + "] "
		}

		if ctx != nil {
			req := ctx.Request
			_, _ = fmt.Fprintf(out, "%v ERROR %s %v[%v %v] -> %v", now, site.Name(), key, ctx.OriginalURL, req.Method, err)
		} else {
			_, _ = fmt.Fprintf(out, "%v ERROR %s %v[%T] -> %v", now, site.Name(), key, scope.Value, err)
		}
	}
}
