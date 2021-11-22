package models

type Config struct {
	MinimumBalance float64            `yaml:"minimum_balance"`
	AssetWeights   map[string]float64 `yaml:"asset_weights,flow"`
}
