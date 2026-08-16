package strategies

import (
	"testing"

	"github.com/rodrigo-brito/ninjabot/strategy"
	"github.com/seguidor777/portfel/internal/localkv"
	"github.com/seguidor777/portfel/internal/models"
)

func TestNewStrategy(t *testing.T) {
	testCases := []struct {
		name         string
		strategyName string
		expectErr    bool
	}{
		{
			name:         "Balancer Strategy",
			strategyName: "Balancer",
			expectErr:    false,
		},
		{
			name:         "Unknown Strategy",
			strategyName: "NonExistentStrategy",
			expectErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &models.Config{
				Strategy: tc.strategyName,
				AssetWeights: map[string]float64{
					"BTCUSDT": 1.0,
				},
			}

			strat, err := New(cfg, nil)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error for strategy %s, got nil", tc.strategyName)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error for strategy %s: %v", tc.strategyName, err)
				}
				if strat == nil {
					t.Fatalf("expected non-nil strategy for %s", tc.strategyName)
				}
			}
		})
	}
}

func TestCustomStrategyRegistration(t *testing.T) {
	Register("CustomTestStrategy", func(config *models.Config, kv *localkv.LocalKV) (strategy.Strategy, error) {
		return NewBalancer(config), nil
	})

	cfg := &models.Config{
		Strategy: "CustomTestStrategy",
		AssetWeights: map[string]float64{
			"BTCUSDT": 1.0,
		},
	}

	strat, err := New(cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error creating registered custom strategy: %v", err)
	}
	if strat == nil {
		t.Fatal("expected non-nil strategy")
	}
}
