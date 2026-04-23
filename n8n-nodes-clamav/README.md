# n8n-nodes-clamav

ClamAV file scanning node for n8n.

## Requirements

- n8n v1.77.0+
- Node.js 18+

## Installation

### Development

1. Clone this repository
2. Install dependencies:
   ```bash
   npm install
   ```
3. Build the project:
   ```bash
   npm run build
   ```
4. Link to n8n:
   ```bash
   cd dist
   npm link
   ```
5. Start n8n in development mode with the linked package

### Production

1. Build: `npm run build`
2. Copy `dist` folder to n8n custom nodes directory
3. Restart n8n

## Usage

### Credentials

1. Go to **Credentials** in n8n
2. Create new credential type **ClamAV API**
3. Configure:
   - **API URL**: Your ClamAV API server URL (e.g., `http://localhost:8080`)
   - **API Key**: Your API key

### Node

1. Add **ClamAV File Scan** node to your workflow
2. Select your ClamAV API credentials
3. Configure:
   - **Input Binary Field**: The name of the binary field containing the file to scan
   - **File Name**: (Optional) Custom file name

### Output

The node returns:

| Field | Type | Description |
|-------|------|-------------|
| `fileId` | string | Unique scan identifier |
| `fileName` | string | Name of scanned file |
| `fileSize` | number | File size in bytes |
| `result` | string | Scan result: `clean`, `infected`, or `error` |
| `threat` | string/null | Detected threat name (if infected) |
| `durationMs` | number | Scan duration in milliseconds |
| `scannedAt` | string | ISO timestamp of scan completion |

## Example Workflow

```
[Webhook] -> [Read Binary Files] -> [ClamAV File Scan] -> [IF] -> [Slack/Email]
                                                      |
                                                      └──> [Continue on error]
```

## License

MIT
