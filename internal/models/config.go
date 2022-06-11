package models

type Config struct {
	Strategy          string             `yaml:"strategy"`
	MinimumBalance    float64            `yaml:"minimum_balance"`
	ExpectedPriceDrop float64            `yaml:"expected_price_drop"`
	AssetWeights      map[string]float64 `yaml:"asset_weights,flow"`
}
