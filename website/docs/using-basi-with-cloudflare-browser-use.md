

You can use basi to connect to Cloudflare's Browser Run sessions automate browser tasks programmatically via CDP. Make sure you create a Account Token with `Browser Rendering - Edit` permission. [Read here](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/) for more


**Prerequisites**:

- basi
- A Cloudflare account with Browser Run enabled
- A Browser Run API token with `Browser Rendering - Edit` permissions

1. Create an env file `.env` to hold the configuration for the Cloudflare endpoint

```sh
#!/bin/sh
export CF_ACCOUNT_ID="<YOUR_CLOUDFLARE_ACCOUNT_ID>";

export CF_API_TOKEN ="<YOUR_CLOUDFLARE_API_TOKEN>";

export CDP_HEADERS='{"Authorization": "Bearer <YOUR_CLOUDFLARE_API_TOKEN>"}'

export URL="wss://api.cloudflare.com/client/v4/accounts/${CF_ACCOUNT_ID}/browser-rendering/devtools/browser?keep_alive=600000"
```

2. Load the environment variables and then run basi.

```sh
source .env

basi run --cdp-endpoint "$URL" test.basi
```

If it succeeds, it means you just ran a Browser test via Basi using a headless browser on Cloudflare's infrastructure. How cool is that!