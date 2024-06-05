import requests
from config import OSMOSIS_QUOTE_API_URL, BTC_DENOM, USDC_DENOM

def get_osmosis_price(token_in, token_out, token_in_amount):
    try:
        url = f"{OSMOSIS_QUOTE_API_URL}?tokenIn={token_in_amount}{token_in}&tokenOutDenom={token_out}&humanDenoms=false"
        response = requests.get(url)
        response.raise_for_status()
        data = response.json()
        amount_out = float(data['amount_out'])
        return amount_out
    except requests.exceptions.RequestException as e:
        print(f"Error fetching price from Osmosis: {e}")
        return None

def get_osmosis_btc_to_usdc_price(token_in_amount):
    return get_osmosis_price(BTC_DENOM, USDC_DENOM, token_in_amount)

def get_osmosis_usdc_to_btc_price(token_in_amount):
    return get_osmosis_price(USDC_DENOM, BTC_DENOM, token_in_amount)

# TODO: Implement Swap Transactions using Block SDK(Top of block auction)