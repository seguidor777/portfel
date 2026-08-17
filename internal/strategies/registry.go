package strategies

import (
	"fmt"

	"github.com/rodrigo-brito/ninjabot/strategy"
	"github.com/seguidor777/portfel/internal/localkv"
	"github.com/seguidor777/portfel/internal/models"
)

type PortfelStrategy interface {
	strategy.Strategy
	GetData() *models.StrategyData
}

type Factory func(config *models.Config, kv *localkv.LocalKV) (PortfelStrategy, error)

var registry = make(map[string]Factory)

// Register registers a strategy constructor under a name.
// Each strategy self-registers in its own init() function.
func Register(name string, factory Factory) {
	registry[name] = factory
}

// New creates a strategy instance dynamically based on config.Strategy.
func New(config *models.Config, kv *localkv.LocalKV) (PortfelStrategy, error) {
	factory, ok := registry[config.Strategy]
	if !ok {
		return nil, fmt.Errorf("invalid strategy: %s", config.Strategy)
	}
	return factory(config, kv)
}
