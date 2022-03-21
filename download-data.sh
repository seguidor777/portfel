#!/bin/bash

set -e

# Change this according your portfolio
pairs=(BTC ADA ETH SOL FTM DOT AVAX LINK LUNA MATIC ROSE MANA SAND AUDIO UNI)
timeframe=1d
days=365

for pair in ${pairs[@]}; do
  ninjabot download --pair ${pair}USDT --timeframe $timeframe --days $days --output testdata/${pair}USDT-${timeframe}.csv
done
