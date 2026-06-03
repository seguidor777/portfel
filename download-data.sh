#!/bin/bash

set -e

# Change this according your portfolio
pairs=(BTC ETH SOL BNB LINK ARB AAVE ADA UNI AVAX POL NEAR SUI DOT)
timeframe=1d
days=365
#start=2021-04-14
#end=2021-11-10

for pair in ${pairs[@]}; do
  ninjabot download --pair ${pair}USDT --timeframe $timeframe --output testdata/${pair}USDT-${timeframe}.csv --days $days #--start $start --end $end
done
