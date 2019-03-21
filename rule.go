package querylog

import (
	"regexp"
	"strings"

	"github.com/aghape/core"

	"github.com/moisespsena-go/aorm"
)

type Actions []string

func (actions Actions) Accept(action string) bool {
	if len(actions) > 0 {
		for i := range actions {
			if actions[i] == action {
				return true
			}
		}
		return false
	}
	return true
}

type TermRule interface {
	Accept(s *aorm.Scope) bool
}

type ContainsRule struct {
	Term string
}

func (r ContainsRule) Accept(s *aorm.Scope) bool {
	return strings.Contains(s.SQL, r.Term)
}

type RegexRule struct {
	Term *regexp.Regexp
}

func (r RegexRule) Accept(s *aorm.Scope) bool {
	return r.Term.MatchString(s.SQL)
}

type Rule struct {
	ConfigRule
	Term TermRule
}

func (r *Rule) Accept(ctx *core.Context, site core.SiteInterface, action string, scope *aorm.Scope) bool {
	if !r.Actions.Accept(action) {
		return false
	}

	if r.Hosts != nil {
		if ctx == nil || ctx.Request == nil || !r.Hosts.Accept(ctx.Request.Host) {
			return false
		}
	}

	return r.Term.Accept(scope)
}
