---

version: "1.0.1"
endpoints:
  bsc01:
    chain: "binance-chain"
    network: mainnet
    url: https://bsc-dataseed1.defibit.io
    headers:
      Authorization: Bearer token119

  bsc02:
    chain: "binance-chain"
    network: mainnet
    url: https://bsc-dataseed2.defibit.io
    
