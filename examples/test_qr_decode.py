#!/usr/bin/env python3
"""
QR Code Content Decoder and Validator
Decodes QR codes from ./output/ and verifies they contain expected content
patterns. Supports PNG decoding via zbar/pyzbar and SVG content inspection.
Usage:
    pip install opencv-python-headless pyzbar pillow
    python3 test_qr_decode.py
"""
import os
import sys
import re
import json
import glob
import time
from pathlib import Path
from typing import Optional, Tuple
# QR content pattern definitions
QR_PATTERNS = {
    "text_hello": {
        "pattern": r"Hello.*World",
        "description": "Text: Hello World",
    },
    "text_quote": {
        "pattern": r"great work.*love what you do",
        "description": "Text: Steve Jobs quote",
    },
    "text_unicode": {
        "pattern": r"unicode.*\u00a9|\u2605|\u2665",
        "description": "Text: Unicode characters",
    },
    "url_google": {
        "pattern": r"https?:
        "description": "URL: Google",
    },
    "url_github": {
        "pattern": r"https?:
        "description": "URL: GitHub",
    },
    "url_gomod": {
        "pattern": r"github\.com/os-gomod/qrcode",
        "description": "URL: Go QR module",
    },
    "wifi_home": {
        "pattern": r"WIFI:T:WPA2;S:MyHomeWiFi",
        "description": "WiFi: Home network",
    },
    "wifi_open": {
        "pattern": r"WIFI:T:nopass;S:CafeFreeWiFi",
        "description": "WiFi: Open network",
    },
    "contact_vcard": {
        "pattern": r"BEGIN:VCARD",
        "description": "vCard contact",
    },
    "contact_mecard": {
        "pattern": r"MECARD:N:",
        "description": "MeCard contact",
    },
    "msg_sms": {
        "pattern": r"smsto:\+?\d+",
        "description": "SMS message",
    },
    "msg_mms": {
        "pattern": r"mms:\+?\d+",
        "description": "MMS message",
    },
    "msg_phone": {
        "pattern": r"tel:\+?\d+",
        "description": "Phone call",
    },
    "email": {
        "pattern": r"mailto:hello@example\.com",
        "description": "Email",
    },
    "geo_sf": {
        "pattern": r"geo:37\.7749,-122\.4194",
        "description": "Geo: San Francisco",
    },
    "maps_google": {
        "pattern": r"maps\.google\.com.*Eiffel\+Tower",
        "description": "Google Maps: Eiffel Tower",
    },
    "maps_directions": {
        "pattern": r"maps\.google\.com.*dir.*Times\+Square",
        "description": "Google Maps: Directions",
    },
    "maps_place": {
        "pattern": r"maps\.google\.com.*Statue\+of\+Liberty",
        "description": "Google Maps: Place",
    },
    "maps_apple": {
        "pattern": r"maps\.apple\.com",
        "description": "Apple Maps",
    },
    "calendar": {
        "pattern": r"BEGIN:VEVENT.*Go Conference",
        "description": "Calendar event",
    },
    "event_ticket": {
        "pattern": r"EVENT-TICKET:EVT-2026-0042",
        "description": "Event ticket",
    },
    "social_twitter": {
        "pattern": r"https?:
        "description": "Twitter profile",
    },
    "social_instagram": {
        "pattern": r"https?:
        "description": "Instagram profile",
    },
    "social_facebook": {
        "pattern": r"https?:
        "description": "Facebook page",
    },
    "social_linkedin": {
        "pattern": r"https?:
        "description": "LinkedIn profile",
    },
    "social_telegram": {
        "pattern": r"https?:
        "description": "Telegram",
    },
    "social_youtube_channel": {
        "pattern": r"https?:
        "description": "YouTube channel",
    },
    "social_youtube_video": {
        "pattern": r"https?:
        "description": "YouTube video",
    },
    "social_spotify_track": {
        "pattern": r"https?:
        "description": "Spotify track",
    },
    "social_spotify_playlist": {
        "pattern": r"https?:
        "description": "Spotify playlist",
    },
    "chat_whatsapp": {
        "pattern": r"https?:
        "description": "WhatsApp chat",
    },
    "chat_zoom": {
        "pattern": r"https?:
        "description": "Zoom meeting",
    },
    "market_playstore": {
        "pattern": r"https?:
        "description": "Google Play Store",
    },
    "market_appstore": {
        "pattern": r"https?:
        "description": "Apple App Store",
    },
    "payment_paypal": {
        "pattern": r"https?:
        "description": "PayPal payment",
    },
    "payment_bitcoin": {
        "pattern": r"bitcoin:bc1q",
        "description": "Bitcoin address",
    },
    "payment_ethereum": {
        "pattern": r"ethereum:0x",
        "description": "Ethereum address",
    },
    "ibeacon": {
        "pattern": r"beacon.*uuid=",
        "description": "iBeacon",
    },
    "ntp": {
        "pattern": r"ntp:
        "description": "NTP server",
    },
    "builder_custom": {
        "pattern": r"github\.com/os-gomod/qrcode",
        "description": "Builder: custom URL",
    },
    "builder_quick": {
        "pattern": r"Quick builder test",
        "description": "Builder: quick test",
    },
    "builder_quickfile": {
        "pattern": r"Builder QuickFile test",
        "description": "Builder: quickfile test",
    },
    "advanced_rounded": {
        "pattern": r"Rounded Modules",
        "description": "Advanced: rounded",
    },
    "advanced_circle": {
        "pattern": r"Circle Modules",
        "description": "Advanced: circle",
    },
    "advanced_gradient": {
        "pattern": r"Gradient QR",
        "description": "Advanced: gradient",
    },
    "advanced_diamond": {
        "pattern": r"Diamond Modules",
        "description": "Advanced: diamond",
    },
}
class QRDecoder:
    """Decodes QR codes from PNG/SVG files."""
    def __init__(self, output_dir: str = "./output"):
        self.output_dir = Path(output_dir)
        self.results = []
        self._zbar_available = None
        self._cv2_available = None
    def _check_zbar(self) -> bool:
        if self._zbar_available is None:
            try:
                from pyzbar.pyzbar import decode as pyzbar_decode
                self._zbar_available = True
            except ImportError:
                self._zbar_available = False
                print("  [WARN] pyzbar not installed. PNG decoding disabled.")
                print("         Install with: pip install pyzbar opencv-python-headless")
        return self._zbar_available
    def _check_cv2(self) -> bool:
        if self._cv2_available is None:
            try:
                import cv2
                self._cv2_available = True
            except ImportError:
                self._cv2_available = False
        return self._cv2_available
    def decode_png(self, filepath: Path) -> Optional[str]:
        """Decode a QR code from a PNG file using pyzbar."""
        if not self._check_zbar():
            return None
        try:
            import cv2
            from pyzbar.pyzbar import decode as pyzbar_decode
            img = cv2.imread(str(filepath))
            if img is None:
                return None
            results = pyzbar_decode(img)
            if results:
                return results[0].data.decode("utf-8", errors="replace")
        except Exception as e:
            print(f"  [ERROR] Failed to decode {filepath.name}: {e}")
        return None
    def decode_svg_content(self, filepath: Path) -> Optional[str]:
        """For SVG files, we cannot decode the visual QR pattern.
        Instead, verify SVG structure validity."""
        try:
            content = filepath.read_text(encoding="utf-8")
            if "<?xml" in content and "<svg" in content and "</svg>" in content:
                return "VALID_SVG"
        except Exception:
            pass
        return None
    def run(self) -> dict:
        """Run verification on all output files."""
        print("=" * 60)
        print("=== QR Code Content Decoder ===")
        print(f"Output directory: {self.output_dir}")
        print()
        if not self.output_dir.exists():
            print("[FAIL] Output directory does not exist!")
            return {"valid": 0, "invalid": 0, "skipped": 0, "results": []}
        png_files = sorted(self.output_dir.glob("*.png"))
        svg_files = sorted(self.output_dir.glob("*.svg"))
        txt_files = sorted(self.output_dir.glob("*.txt"))
        valid = 0
        invalid = 0
        skipped = 0
        # Decode PNGs
        print(f"--- PNG Files ({len(png_files)} found) ---")
        for f in png_files:
            base = f.stem
            decoded = self.decode_png(f)
            if decoded is None:
                if not self._check_zbar():
                    skipped += 1
                    print(f"  [SKIP] {f.name} (decoder not available)")
                else:
                    invalid += 1
                    print(f"  [FAIL] {f.name} - could not decode")
                continue
            pattern_info = QR_PATTERNS.get(base)
            if pattern_info:
                match = re.search(pattern_info["pattern"], decoded, re.IGNORECASE | re.DOTALL)
                if match:
                    valid += 1
                    preview = decoded[:80].replace("\n", "\\n")
                    print(f"  [PASS] {f.name} - {pattern_info['description']}")
                    print(f"         Content: {preview}...")
                else:
                    invalid += 1
                    print(f"  [WARN] {f.name} - decoded but pattern not matched")
                    print(f"         Expected: {pattern_info['description']}")
                    print(f"         Got: {decoded[:100]}")
            else:
                valid += 1
                preview = decoded[:60].replace("\n", "\\n")
                print(f"  [PASS] {f.name} - decoded OK: {preview}...")
        # Validate SVGs
        print(f"\n--- SVG Files ({len(svg_files)} found) ---")
        for f in svg_files:
            base = f.stem
            result = self.decode_svg_content(f)
            if result == "VALID_SVG":
                valid += 1
                # Check size
                size = f.stat().st_size
                print(f"  [PASS] {f.name} - valid SVG ({size} bytes)")
            else:
                invalid += 1
                print(f"  [FAIL] {f.name} - invalid SVG")
        # Validate TXT files
        print(f"\n--- TXT Files ({len(txt_files)} found) ---")
        for f in txt_files:
            base = f.stem
            if f.stat().st_size > 0:
                valid += 1
                content = f.read_text()[:80].replace("\n", "\\n")
                print(f"  [PASS] {f.name} - non-empty ({len(content)}+ chars)")
            else:
                invalid += 1
                print(f"  [FAIL] {f.name} - empty file")
        # Summary
        total = valid + invalid + skipped
        print()
        print("=" * 60)
        print("=== Decoder Summary ===")
        print(f"  Total files:   {total}")
        print(f"  Valid:         {valid}")
        print(f"  Invalid:       {invalid}")
        print(f"  Skipped:       {skipped}")
        if skipped > 0:
            print(f"  (Skipped due to missing decoder - install pyzbar)")
        print("=" * 60)
        all_ok = invalid == 0
        if skipped > 0 and invalid == 0:
            print("Result: PASSED (with skipped files)")
        elif all_ok:
            print("Result: ALL CHECKS PASSED")
        else:
            print("Result: SOME CHECKS FAILED")
        return {"valid": valid, "invalid": invalid, "skipped": skipped}
def main():
    output_dir = os.environ.get("QR_OUTPUT_DIR", "./output")
    if len(sys.argv) > 1:
        output_dir = sys.argv[1]
    decoder = QRDecoder(output_dir)
    result = decoder.run()
    if result["invalid"] > 0:
        sys.exit(1)
    sys.exit(0)
if __name__ == "__main__":
    main()
