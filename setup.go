package querylog

import (
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/moisespsena-go/path-helpers"

	"github.com/moisespsena-go/aorm"

	"github.com/ecletus/ecletus"
	"github.com/ecletus/core"
)

type Output struct {
	io.WriteCloser
	Rules  []*Rule
	closer bool
}

func (o Output) Close() error {
	if o.closer {
		return o.WriteCloser.Close()
	}
	return nil
}

func initOutputs(cfg *Config) (outputs []*Output) {
	var (
		err error
		m   = map[string]io.WriteCloser{}
	)

	defer func() {
		if err != nil {
			for _, o := range m {
				_ = o.Close()
			}
			panic(err)
		}
	}()

	for _, out := range cfg.Outputs {
		if out.Dest == "stderr" {
			out.Dest = ""
		}

		w := m[out.Dest]
		if w == nil {
			if out.Dest == "" {
				w = os.Stderr
			} else {
				for out.Truncate {
					if _, err = os.Stat(out.Dest); err == nil {
						if err = os.Truncate(out.Dest, 0); err != nil {
							return
						}
					} else {
						return
					}
				}

				if out.Perm == 0 {
					out.Perm = 0600
				}

				var ok bool
				if w, ok = m[out.Dest]; !ok {
					if err = path_helpers.MkdirAllIfNotExists(filepath.Dir(out.Dest)); err != nil {
						return
					}
					w, err = os.OpenFile(out.Dest, os.O_RDWR|os.O_APPEND|os.O_CREATE, out.Perm)
					if err != nil {
						return
					}
				}
			}
		}

		o := &Output{WriteCloser: w, closer: out.Dest != ""}

		if _, ok := m[out.Dest]; ok {
			o.closer = false
		} else {
			m[out.Dest] = w
		}

		for _, r := range out.Rules {
			rule := &Rule{ConfigRule: r}
			if r.Regex {
				regex := RegexRule{}
				if regex.Term, err = regexp.Compile(r.Term); err != nil {
					return
				}
				rule.Term = regex
			} else {
				rule.Term = &ContainsRule{r.Term}
			}
			o.Rules = append(o.Rules, rule)
		}

		outputs = append(outputs, o)
	}
	return
}

func Setup(cfg *Config, agp *ecletus.Ecletus, sites *core.SitesRegister) {
	var (
		outputs = initOutputs(cfg)
	)

	sites.OnAdd(func(site *core.Site) {
		for _, o := range outputs {
			for _, r := range o.Rules {
				if r.Sites.Accept(site.Name()) {
					_ = site.EachDB(func(db *core.DB) (err error) {
						if r.Dbs.Accept(db.Name) {
							db.InitCallback(func(DB *core.DB) {
								for _, table := range r.Tables {
									if r.Error {
										db.DB = db.DB.ScopeErrorCallback(loggerError(o, r, site))
									} else {
										var lgs *aorm.ScopeLoggers
										db.DB, lgs, _ = db.DB.Loggers(table, true)
										if len(r.Actions) == 0 {
											r.Actions = append(r.Actions, aorm.LOG_EXEC)
										}
										for _, action := range r.Actions {
											lgs.Register(action, logger(o, r, site))
										}
									}
								}
							})
						}
						return nil
					})
				}
			}
		}
	})

	agp.Done(func() {
		for _, out := range outputs {
			_ = out.Close()
		}
	})
}
