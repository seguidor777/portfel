#!/bin/bash

set -e

# Change this according your portfolio
pairs=(BTC ADA ETH SOL BNB DOT AVAX LINK FTM MATIC ROSE MANA SAND GALA AUDIO)
timeframe=1d
days=365

for pair in ${pairs[@]}; do
  ninjabot download --pair ${pair}BUSD --timeframe $timeframe --days $days --output testdata/${pair}BUSD-${timeframe}.csv
done
