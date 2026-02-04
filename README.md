# poster

poster is a Go CLI for posting photos, reels, and carousels to Instagram using the official Graph API. It uploads local media to https://uguu.se for temporary hosting, then publishes via the Instagram Graph API.

## Installation

### Go install

```bash
go install github.com/mahmoudashraf93/poster/cmd/poster@latest
```

### Build from source

```bash
make build
./bin/poster --help
```

## Setup (Meta + Instagram)

### Prerequisites

1. Your Instagram account must be a **Business** or **Creator** account.
2. The Instagram account must be linked to a **Facebook Page**.

### Create the Meta app

1. Create a Meta app in the [Meta for Developers](https://developers.facebook.com/) dashboard (Business app type is usually the right choice for posting).
2. Add the [Instagram Graph API](https://developers.facebook.com/docs/instagram-api/) product.
3. Add [Facebook Login](https://developers.facebook.com/docs/facebook-login/) (required to generate User access tokens).
4. If you plan to list pages by Business Manager, also add **Business Management**.

### Roles + App Mode

1. In Development Mode, only app admins/developers/testers can generate tokens.
2. Add yourself as an app admin/developer/tester and accept the invitation.
3. If you want to use the app with accounts that are not admins/testers, you must switch to Live Mode and complete [App Review](https://developers.facebook.com/docs/app-review/) for the required permissions.

### Required permissions (scopes)

Minimum for posting:

- `instagram_basic`
- `instagram_content_publish`
- `pages_show_list`

Common extras (depending on what you call):

- `pages_read_engagement` (often required to read Page details that link to Instagram)
- `business_management` (required for `poster owned-pages` and Business Manager page discovery)

### Get a short‑lived user token (Graph API Explorer)

1. Open **Tools →** [Graph API Explorer](https://developers.facebook.com/tools/explorer/) in the Meta Developers dashboard.
2. Select your app from the app dropdown.
3. Click **Get Token → Get User Access Token**.
4. Check the permissions you need (see list above), then authorize.
5. Copy the generated **short‑lived** token.
6. (Optional) Use [Access Token Debugger](https://developers.facebook.com/tools/debug/accesstoken/) to confirm the granted scopes.

### Exchange for a long‑lived token

```bash
poster token exchange --short-token "<short_token>"
```

### Fetch IDs used by the CLI

1. Get `IG_PAGE_ID` (from Page settings or `poster owned-pages`).
2. Run:

```bash
poster --page-id <PAGE_ID> account
```

3. It prints `IG_USER_ID=...` for your Instagram Business/Creator account.

### Business Manager notes

1. If the Page is owned by a Business Manager, grab the Business ID from [Business Settings](https://business.facebook.com/settings/) → Business Info.
2. Set `IG_BUSINESS_ID` and use:

```bash
poster owned-pages --business-id <BUSINESS_ID>
```

3. Ensure your app has the `business_management` permission and your user has admin rights on the Business and Page.

## Usage

Global flags:

- `--profile`: Profile name (selects stored config + keyring token).
- `--user-id`: Override `IG_USER_ID` at runtime.
- `--page-id`: Override `IG_PAGE_ID` at runtime.
- `--business-id`: Override `IG_BUSINESS_ID` at runtime.

Do you need BOTH `IG_USER_ID` and `IG_PAGE_ID` to post?

- For posting, you only need `IG_USER_ID` + `IG_ACCESS_TOKEN`.
- `IG_PAGE_ID` is only needed to lookup the IG user ID (via `poster account`).

So the typical flow is:

1. Get `IG_PAGE_ID` (one time).
2. Run:

```bash
poster --page-id <PAGE_ID> account
```

3. It prints `IG_USER_ID=...`
4. From then on, use `IG_USER_ID` (you can keep `IG_PAGE_ID` around, but it’s not required for posting).

### Post a photo

```bash
poster photo --file path/to/photo.jpg --caption "hello"
```

```bash
poster photo --url https://example.com/photo.jpg --caption "hello"
```

### Post a reel

```bash
poster reel --file path/to/video.mp4 --caption "hello"
```

### Post a carousel

```bash
poster carousel --files img1.jpg img2.jpg --caption "hello"
```

### Token utilities

```bash
poster token exchange --short-token "<short_token>"
poster token debug
```

### Account utilities

```bash
poster account
poster owned-pages --business-id <BUSINESS_ID>
```

### Profile management (keyring-backed)

Profiles store non-secret values in `~/.config/poster/config.json`, while access tokens are stored in the OS keyring (Keychain, Secret Service, or encrypted file backend depending on configuration).

Resolution order:

1. CLI flags (e.g. `--user-id`, `--page-id`, `--business-id`)
2. Selected profile (`--profile` or `IG_PROFILE`)
3. Environment variables / `.env`

```bash
poster profile set brand-a --access-token "<token>" --user-id <IG_USER_ID> --page-id <PAGE_ID> --business-id <BUSINESS_ID>
poster profile show brand-a
poster profile list
poster profile delete brand-a
```

### Keyring backend (keychain vs encrypted file)

Backends:

- `auto` (default): picks the best backend for the platform.
- `keychain`: macOS Keychain (recommended on macOS).
- `file`: encrypted on-disk keyring (requires a password).

Set backend (writes `keyring_backend` into `config.json`):

```bash
poster keyring file
poster keyring keychain
poster keyring auto
```

Show current backend + source:

```bash
poster keyring
```

Non-interactive runs (CI/ssh): file backend requires `POSTER_KEYRING_PASSWORD`.

```bash
export POSTER_KEYRING_PASSWORD='...'
```

Force backend via env (overrides config):

```bash
export POSTER_KEYRING_BACKEND=file
```

## Environment variables

Set these in `.env` (see `.env.example`) or export them in your shell.

- `IG_APP_ID`: Meta app ID.
- `IG_APP_SECRET`: Meta app secret.
- `IG_ACCESS_TOKEN`: Long-lived Instagram User access token.
- `IG_PROFILE`: Profile name (default: `default`).
- `IG_PAGE_ID`: Facebook Page ID connected to your Instagram account.
- `IG_BUSINESS_ID`: Meta Business ID (for listing owned pages).
- `IG_USER_ID`: Instagram Business/User ID.
- `IG_GRAPH_VERSION`: Graph API version (default: `v19.0`).
- `IG_POLL_INTERVAL`: Polling interval for media processing (default: `5s`).
- `IG_POLL_TIMEOUT`: Polling timeout for media processing (default: `300s`).
- `POSTER_KEYRING_BACKEND`: Keyring backend (`auto`, `keychain`, `file`). Overrides config.
- `POSTER_KEYRING_PASSWORD`: Password for encrypted file backend (use in non-interactive runs).

## Token refresh

Long-lived tokens expire. When yours is near expiry:

1. Generate a new short-lived token via **Tools → Graph API Explorer**.
2. Run:

```bash
poster token exchange --short-token "<short_token>"
```

3. Update `IG_ACCESS_TOKEN` with the new value.
