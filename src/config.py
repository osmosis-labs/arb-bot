import os
from dotenv import load_dotenv

load_dotenv()

BINANCE_API_KEY = os.getenv('BINANCE_API_KEY')
BINANCE_SECRET_KEY = os.getenv('BINANCE_SECRET_KEY')

OSMOSIS_QUOTE_API_URL = "https://sqs.osmosis.zone/router/quote"
BTC_DENOM = "factory/osmo1z0qrq605sjgcqpylfl4aa6s90x738j7m58wyatt0tdzflg2ha26q67k743/wbtc"
USDC_DENOM = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"

# this is in human readable exponent
DEFAULT_ARB_AMT = "1"

RISK_FACTOR = "0.9"