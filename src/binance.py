import ccxt
from config import BINANCE_API_KEY, BINANCE_SECRET_KEY

binance = ccxt.binance(config={
    'apiKey': BINANCE_API_KEY,
    'secret': BINANCE_SECRET_KEY,
})
def get_binance_btc_to_usdt_price():
    ticker = binance.fetch_ticker(symbol='BTC/USDT')
    return ticker['last']

def get_binance_usdc_to_btc_price():
    btc_price = get_binance_btc_to_usdt_price()
    if btc_price != 0:
        return 1 / btc_price
    else:
        raise ValueError("BTC price from Binance is zero, cannot compute USDC to BTC price")

def get_btc_balance():
    balance = binance.fetch_balance()
    return balance['USDT']['free'], balance['BTC']['free']


def buy_btc_on_binance(amount):
    try:
        order = binance.create_market_buy_order('BTC/USDT', amount)
        print(f"Buy order placed: {order}")
    except Exception as e:
        print(f"Error placing buy order: {e}")

def sell_btc_on_binance(amount):
    try:
        order = binance.create_market_sell_order('BTC/USDT', amount)
        print(f"Sell order placed: {order}")
    except Exception as e:
        print(f"Error placing sell order: {e}")