import hashlib
import logging
import time
from collections import deque
from typing import Deque

from bs4 import BeautifulSoup
import requests  # type: ignore

from parser_app.logging_config import configure_logging
from parser_app.ttsiigh_utils import add_schedule_from_xlsx


SCHEDULE_PAGE_URL = "https://tci72.ru/students/schedule/"
BASE_URL = "https://tci72.ru"
REQUEST_TIMEOUT = 10
DOWNLOAD_TIMEOUT = 30
POLL_INTERVAL_SECONDS = 60
MAX_HASHES = 10

configure_logging()
log = logging.getLogger("parser_app")

last_hashes: Deque[str] = deque(maxlen=MAX_HASHES)


def get_download_link() -> str | None:
    try:
        response = requests.get(SCHEDULE_PAGE_URL, timeout=REQUEST_TIMEOUT)
        response.raise_for_status()
    except requests.RequestException:
        log.error("Failed to load schedule page %s", SCHEDULE_PAGE_URL, exc_info=True)
        return None

    try:
        soup = BeautifulSoup(response.text, "html.parser")
        li = soup.find("li", class_="file-download")
        if not li:
            log.error("Failed to find 'file-download' list element on schedule page")
            return None

        link_element = li.find("a")
        if not link_element:
            log.error("Failed to find link inside 'file-download' element")
            return None

        href = link_element.get("href")
        if not href:
            log.error("Found link element without href attribute")
            return None

        full_link = BASE_URL + href
        return full_link
    except Exception:
        log.error("Failed to parse download link from schedule page", exc_info=True)
        return None


def start_listing() -> None:
    log.info("Worker started, monitoring schedule updates from %s...", SCHEDULE_PAGE_URL)
    while True:
        try:
            link = get_download_link()
            if not link:
                log.error("Failed to find download link on the website %s", SCHEDULE_PAGE_URL)
                time.sleep(POLL_INTERVAL_SECONDS)
                continue

            try:
                response = requests.get(link, timeout=DOWNLOAD_TIMEOUT)
                response.raise_for_status()
            except requests.RequestException:
                log.error("Failed to download schedule file from %s", link, exc_info=True)
                time.sleep(POLL_INTERVAL_SECONDS)
                continue

            hash_md5 = hashlib.md5(response.content).hexdigest()
            if hash_md5 in last_hashes:
                time.sleep(POLL_INTERVAL_SECONDS)
                continue

            try:
                result = add_schedule_from_xlsx(response.content)
                if result:
                    log.info("New schedule added successfully: %s", link)
                else:
                    log.info("Schedule from site is not added: %s", link)
            except Exception:
                log.error("Error when processing given schedule %s", link, exc_info=True)

            last_hashes.append(hash_md5)

        except Exception:
            log.error("Unexpected error in worker loop", exc_info=True)

        time.sleep(POLL_INTERVAL_SECONDS)


if __name__ == "__main__":
    start_listing()
