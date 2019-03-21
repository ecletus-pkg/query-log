package querylog

import (
	"fmt"
	"time"

	"github.com/aghape/core"
	"github.com/moisespsena-go/aorm"
)

func logger(out *Output, r *Rule, site core.SiteInterface) func(action string, scope *aorm.Scope) {
	return func(action string, scope *aorm.Scope) {
		ctx := core.ContextFromDB(scope.DB())
		if !r.Accept(ctx, site, action, scope) {
			return
		}

		now := scope.ExecTime
		if now.IsZero() {
			now = time.Now()
		}
		
		var v string
		if r.Values {
			v = aorm.SQLToString(scope.SQL, scope.SQLVars...)
		} else {
			v = scope.SQL
		}

		var key = r.Key
		if key != "" {
			key = "[" + key + "] "
		}

		if ctx != nil {
			req := ctx.Request
			_, _ = fmt.Fprintf(out, "%v %s %v[%v %v] %v -> %v\n", now, site.Name(), key, req.URL, req.Method, action, v)
		} else {
			_, _ = fmt.Fprintf(out, "%v %s %v[%T] %v -> %v\n", now, site.Name(), key, scope.Value, action, v)
		}
		return
	}
}

func loggerError(out *Output, r *Rule, site core.SiteInterface) func(scope *aorm.Scope, err error) {
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
