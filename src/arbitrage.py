import time
from binance import get_binance_usdc_to_btc_price, get_binance_btc_to_usdt_price, get_btc_balance, buy_btc_on_binance
from osmosis import get_osmosis_btc_to_usdc_price
from config import DEFAULT_ARB_AMT, RISK_FACTOR

osmosis_wbtc_exponent = 8
osmosis_usdc_exponent = 6
def check_arbitrage():
    binance_btc_price = get_binance_btc_to_usdt_price()
    print(binance_btc_price)

    binance_usdc_price = get_binance_usdc_to_btc_price()
    print(binance_usdc_price)

    balance = get_btc_balance()

    print(balance)
    # First get a brief approx on price on Osmosis
    # We call this an approximation since actual quote can differ depending on amount & route
    osmosis_btc_price = get_osmosis_btc_to_usdc_price(int(DEFAULT_ARB_AMT) * pow(10, osmosis_wbtc_exponent))

    # If (binance price < osmosis BTC price * risk factor),
    # buy BTC on Binanace, sell BTC on Osmosis
    if (binance_btc_price < osmosis_btc_price * RISK_FACTOR ):
        print("here")
    elif (binance_btc_price * RISK_FACTOR > osmosis_btc_price ):
        print("here")

def main():
    while True:
        check_arbitrage()
        time.sleep(60)  # Adjust the interval as needed


if __name__ == "__main__":
    main()