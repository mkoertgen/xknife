# X-Knife

A command-line tool for Twitter/X API v2 that focuses on bot detection and user analysis. X-Knife helps you analyze Twitter accounts and detect potentially automated behavior through various metrics and algorithms.

## Features

- **User Information Retrieval**: Get detailed information about any Twitter user including account creation date, verification status, follower counts, and more
- **Follower Analysis**: Analyze recent followers of any account
- **Bot Detection Scoring**: Advanced scoring algorithm that evaluates accounts based on:
  - Verification status and account protection
  - Follower-to-following ratios with mathematical scoring
  - Account age vs. following velocity (detects rapid following behavior)
  - Additional metrics for comprehensive bot detection

## Installation

### Prerequisites

- Go 1.23.2 or later
- Twitter API v2 Bearer Token (OAuth 2.0)

### Building from Source

```bash
git clone <repository-url>
cd xknife/src/go/xknife
go build -o xknife
```

### Using Goreleaser

The project is configured for Goreleaser builds. Binary releases will include version information.

## Configuration

### Environment Variables

Set up your Twitter API credentials by copying the sample environment file:

```bash
cp src/go/.env.sample .env
```

Then edit the `.env` file with your credentials:

```bash
export GOTWI_API_KEY=your-api-key
export GOTWI_API_KEY_SECRET=your-api-key-secret
export GOTWI_ACCESS_TOKEN=your-oauth-access-token
export GOTWI_ACCESS_TOKEN_SECRET=your-oauth-access-token-secret
```

### Config File

X-Knife supports configuration via a YAML file located at `$HOME/.xknife.yaml`. You can also specify a custom config file using the `--config` flag.

## Usage

### Basic Commands

Get information about a specific user:

```bash
xknife get --user username
```

Get recent followers for a user:

```bash
xknife followers --user username --size 50
```

### Command Options

- `--user, -u`: Specify the Twitter username (default: "mkoertg")
- `--id`: Specify the Twitter user ID directly
- `--size, -s`: Number of items to retrieve (default: 20)
- `--config, -c`: Path to config file (default: $HOME/.xknife.yaml)

### Examples

Analyze a user's account:

```bash
xknife get --user elonmusk
```

Check recent followers with custom page size:

```bash
xknife followers --user someuser --size 100
```

## Bot Detection Algorithm

The bot detection scoring system evaluates accounts on a 0-100% scale:

1. **Verified & Protected Accounts**: Automatically receive 100% (human) score
2. **Follower/Following Ratio**: Uses arctangent function to score based on followers vs. following
3. **Following Velocity**: Analyzes daily following rate since account creation (accounts following >20 users/day average are flagged as suspicious)

The algorithm combines multiple factors to provide a comprehensive bot probability score.

## Dependencies

- [michimani/gotwi](https://github.com/michimani/gotwi) - Twitter API v2 Go library
- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [spf13/viper](https://github.com/spf13/viper) - Configuration management
- [carlmjohnson/versioninfo](https://github.com/carlmjohnson/versioninfo) - Version information

## Development

The project structure:

```
src/go/xknife/
├── main.go           # Entry point with version handling
├── cmd/
│   ├── root.go       # CLI root command and configuration
│   └── bots.go       # User analysis and bot detection commands
├── go.mod            # Go module definition
└── .goreleaser.yaml  # Release configuration
```

## Links & References

### API Documentation

- [X API v2 Documentation](https://developer.x.com/en/docs/x-api)
- [gotwi Library](https://github.com/michimani/gotwi)

### Research & Tools

- [Botometer X](https://botometer.osome.iu.edu/)
- [A Decade of Social Bot Detection](https://cacm.acm.org/research/a-decade-of-social-bot-detection/)
- [Twibot-22](https://twibot22.github.io/)
- [GeoGebra Graphing Calculator](https://www.geogebra.org/graphing) (used for algorithm visualization)

## License

MIT License - see [LICENSE](LICENSE) file for details.
