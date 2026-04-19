# QR Code Library - Example & Verification

A comprehensive example demonstrating all features of the `github.com/os-gomod/qrcode` Go library, with verification scripts to validate generated output.

## Prerequisites

- Go 1.26.1+
- (Optional) Python 3 with `pyzbar`, `opencv-python-headless`, `pillow` for QR code decoding

## Project Structure

```
example/
  go.mod                 # Go module (uses local replace directive)
  example.go             # Main example - generates 45+ QR codes
  verify_compilation.sh  # Compilation verification
  verify_qr_codes.sh     # File format and size verification
  test_qr_decode.py      # Python QR code content decoder
  verify_all.sh          # Master verification script
  README.md              # This file
  output/                # Generated QR codes (created on first run)
```

## Quick Start

```bash
# 1. Compile and verify
chmod +x verify_*.sh
./verify_compilation.sh

# 2. Generate QR codes
go run example.go

# 3. Verify generated files
./verify_qr_codes.sh

# 4. Decode and validate content (optional)
pip install opencv-python-headless pyzbar pillow
python3 test_qr_decode.py

# 5. Run full verification suite
./verify_all.sh
```

## Generated QR Code Types

| Category | Files | Formats |
|----------|-------|---------|
| Text | `text_hello`, `text_quote`, `text_unicode` | PNG + SVG |
| URL | `url_google`, `url_github`, `url_gomod` | PNG |
| WiFi | `wifi_home`, `wifi_open`, `wifi_special` | PNG |
| Contact | `contact_vcard`, `contact_mecard` | PNG + SVG |
| Messaging | `msg_sms`, `msg_mms`, `msg_phone` | PNG |
| Email | `email` | PNG |
| Geo/Maps | `geo_sf`, `maps_google`, `maps_directions`, `maps_place`, `maps_apple` | PNG |
| Calendar | `calendar`, `event_ticket` | PNG |
| Social | Twitter, Instagram, Facebook, LinkedIn, Telegram, YouTube, Spotify (9 files) | PNG |
| Chat | `chat_whatsapp`, `chat_zoom` | PNG |
| Market | `market_playstore`, `market_appstore` | PNG |
| Payments | `payment_paypal`, `payment_bitcoin`, `payment_ethereum` | PNG |
| Other | `ibeacon`, `ntp` | PNG |
| Builder | `builder_custom`, `builder_quick`, `builder_quickfile` | PNG |
| Advanced | `advanced_rounded`, `advanced_circle`, `advanced_gradient`, `advanced_diamond` | PNG |
| Batch | `batch_text`, `batch_url`, `batch_hello` | PNG |
| Other formats | `terminal.txt`, `base64.txt` | TXT |
| Edge cases | `edge_long_text`, `edge_empty`, `edge_special_chars` | PNG |

## Verification Scripts

### verify_compilation.sh
Checks that the Go code compiles successfully:
```bash
./verify_compilation.sh   # Exit 0 = pass, 1 = fail
```

### verify_qr_codes.sh
Validates generated files:
- Output directory exists
- At least 30 files generated
- No zero-byte files
- PNG files have valid PNG headers
- SVG files contain valid XML/SVG markup
- TXT files are non-empty

### test_qr_decode.py
Decodes QR codes and checks content patterns using `pyzbar`:
```bash
python3 test_qr_decode.py [output_dir]
```

### verify_all.sh
Runs all verification steps in sequence:
```bash
./verify_all.sh
```

## Library API Demonstrated

- **Generator**: `qrcode.New(opts...)` with caching, error correction, sizing
- **Quick helpers**: `Quick()`, `QuickSVG()`, `QuickFile()`, etc.
- **Builder pattern**: `NewBuilder().Size(300).ErrorCorrection(LevelH).Build()`
- **All payload types**: Text, URL, WiFi, vCard, MeCard, SMS, MMS, Phone, Email, Geo, Calendar, Event, Social media, WhatsApp, Zoom, Market, PayPal, Crypto, iBeacon, NTP
- **Formats**: PNG, SVG, Terminal, PDF, Base64
- **Advanced rendering**: Rounded modules, circle modules, gradient, diamond modules
- **Batch processing**: Concurrent batch generation with stats
- **Edge cases**: Empty data, long text, special characters, context cancellation
