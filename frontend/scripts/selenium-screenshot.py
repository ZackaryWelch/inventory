#!/usr/bin/env python3
"""
Proper screenshot capture using Selenium with real waits
"""

import sys
import time
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import TimeoutException

def capture_screenshot(url, output_path, wait_seconds=20, description=""):
    """Capture screenshot with proper waiting for app to load"""

    print(f"\n{description}")
    print(f"  URL: {url}")
    print(f"  Wait: {wait_seconds}s for app to fully render")
    print(f"  Output: {output_path}")

    # Configure Chrome options
    chrome_options = Options()
    chrome_options.add_argument('--headless')
    chrome_options.add_argument('--no-sandbox')
    chrome_options.add_argument('--disable-dev-shm-usage')
    chrome_options.add_argument('--disable-gpu')
    chrome_options.add_argument('--window-size=1920,1080')

    driver = None
    try:
        # Create driver
        driver = webdriver.Chrome(options=chrome_options)

        # Navigate to URL
        print(f"  Loading page...")
        driver.get(url)

        # Wait for initial page load
        WebDriverWait(driver, 10).until(
            lambda d: d.execute_script("return document.readyState") == "complete"
        )
        print(f"  Page loaded (readyState=complete)")

        # Additional wait for WASM/React to render
        print(f"  Waiting {wait_seconds}s for app rendering...")
        time.sleep(wait_seconds)

        # Take screenshot
        driver.save_screenshot(output_path)
        print(f"  ✓ Screenshot saved to {output_path}")

        return True

    except Exception as e:
        print(f"  ✗ Error: {e}")
        return False

    finally:
        if driver:
            driver.quit()

def main():
    if len(sys.argv) < 3:
        print("Usage: selenium-screenshot.py <url> <output_path> [wait_seconds] [description]")
        sys.exit(1)

    url = sys.argv[1]
    output = sys.argv[2]
    wait = int(sys.argv[3]) if len(sys.argv) > 3 else 20
    desc = sys.argv[4] if len(sys.argv) > 4 else f"Capturing {url}"

    success = capture_screenshot(url, output, wait, desc)
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()
