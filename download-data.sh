#!/bin/bash

set -e

# Change this according your portfolio
pairs=(BTC ETH ADA SOL FTM DOT AVAX LINK LUNA MATIC ROSE MANA SAND AUDIO UNI ALGO NEAR)
timeframe=1d
days=365

for pair in ${pairs[@]}; do
  ninjabot download --pair ${pair}USDT --timeframe $timeframe --days $days --output testdata/${pair}USDT-${timeframe}.csv
done
