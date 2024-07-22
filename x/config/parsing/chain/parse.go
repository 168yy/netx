package chain

import (
	"github.com/168yy/netx/core/chain"
	"github.com/168yy/netx/core/hop"
	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/core/metadata"
	"github.com/168yy/netx/x/app"
	xchain "github.com/168yy/netx/x/chain"
	"github.com/168yy/netx/x/config"
	hopparser "github.com/168yy/netx/x/config/parsing/hop"
	mdx "github.com/168yy/netx/x/metadata"
)

func ParseChain(cfg *config.ChainConfig, log logger.ILogger) (chain.IChainer, error) {
	if cfg == nil {
		return nil, nil
	}

	chainLogger := log.WithFields(map[string]any{
		"kind":  "chain",
		"chain": cfg.Name,
	})

	var md metadata.IMetaData
	if cfg.Metadata != nil {
		md = mdx.NewMetadata(cfg.Metadata)
	}

	c := xchain.NewChain(cfg.Name,
		xchain.MetadataChainOption(md),
		xchain.LoggerChainOption(chainLogger),
	)

	for _, ch := range cfg.Hops {
		var hop hop.IHop
		var err error

		if ch.Nodes != nil || ch.Plugin != nil {
			if hop, err = hopparser.ParseHop(ch, log); err != nil {
				return nil, err
			}
		} else {
			hop = app.Runtime.HopRegistry().Get(ch.Name)
		}
		if hop != nil {
			c.AddHop(hop)
		}
	}

	return c, nil
}
