#!/bin/bash

set -e

# Change this according your portfolio
pairs=(BTC ETH SOL BNB LINK ARB AAVE ADA UNI AVAX POL NEAR SUI DOT)
timeframe=1d
#days=365
start=2022-06-14
end=2024-03-09

for pair in ${pairs[@]}; do
  ninjabot download --pair ${pair}USDT --timeframe $timeframe --output testdata/${pair}USDT-${timeframe}.csv -start $start --end $end #--days $days
done
