package querylog

import (
	"github.com/aghape/aghape"
	"github.com/aghape/cli"
	"github.com/aghape/core"
	"github.com/aghape/plug"
	"github.com/moisespsena/go-default-logger"
	"github.com/moisespsena/go-error-wrap"
	"github.com/moisespsena/go-path-helpers"
)

const CFG_FILE = "query-log.yaml"

var log = defaultlogger.NewLogger(path_helpers.GetCalledDir())

type Plugin struct {
	plug.EventDispatcher
	ConfigDirKey   string
	SitesReaderKey string
	cfg            *Config
	cfgFile        string
}

func (p *Plugin) RequireOptions() []string {
	return []string{p.ConfigDirKey}
}

func (p *Plugin) OnRegister(options *plug.Options) {
	cli.OnRegister(p, func(e *cli.RegisterEvent) {
		p.cfgFile = options.GetInterface(p.ConfigDirKey).(*aghape.ConfigDir).Path(CFG_FILE)
	})

	p.On(plug.E_POST_INIT, func(e plug.PluginEventInterface) (err error) {
		if p.cfg != nil {
			agp := options.GetInterface(aghape.AGHAPE).(*aghape.Aghape)
			Sites := options.GetInterface(p.SitesReaderKey).(core.SitesReaderInterface)
			Setup(p.cfg, agp, Sites)
		}
		return nil	
	})
}

func (p *Plugin) Init(options *plug.Options) {
	p.cfg = p.loadConfig(options)
}

func (p *Plugin) loadConfig(options *plug.Options) (cfg *Config) {
	configDir := options.GetInterface(p.ConfigDirKey).(*aghape.ConfigDir)
	if ok, err := configDir.Exists(CFG_FILE); err != nil {
		log.Error(errwrap.Wrap(err, "Stat of %q", configDir.Path(CFG_FILE)))
		return nil
	} else if !ok {
		return
	}

	cfg = &Config{}
	if err := configDir.Load(cfg, CFG_FILE); err != nil {
		log.Error(err)
		return nil
	}

	return cfg
}
