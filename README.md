Telegram bot for own needs.

## Required environment variables
- ALLOWED_USERS - list of allowed users in json format (example: `[{"id": 123456789}]`)
- BOT_TOKEN - token from [@BotFather](https://t.me/BotFather)
- GOOGLE_CREDENTIALS - api credentials in json format (see https://console.cloud.google.com/apis/credentials)
- GOOGLE_SPREADSHEET - spreadsheet id

## Supported commands
- status - check bot status
- currency - chech current exchange rate
- meters - set electricity/water meters
- lastmeters - show last electricity/water meters
- whoami - show info about requesting user

## Build
```sh
go build .
```
