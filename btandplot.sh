#!/bin/bash

set -e

# Change this according your portfolio
pairs=(BTC ETH ADA SOL DOT AVAX LINK LUNA MANA MATIC ROSE SAND AUDIO AXS SHIB THETA)
timeframe=1d
days=365

for pair in ${pairs[@]}; do
  ninjabot download --pair ${pair}USDT --timeframe $timeframe --days $days --output testdata/${pair}USDT-${timeframe}.csv
done
