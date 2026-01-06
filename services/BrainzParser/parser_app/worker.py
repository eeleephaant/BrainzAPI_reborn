import hashlib
import logging
import time

from bs4 import BeautifulSoup
import requests  # type: ignore

from parser_app.logging_config import configure_logging
from parser_app.ttsiigh_utils import add_schedule_from_xlsx

configure_logging()
log = logging.getLogger("parser_app")

last_hashes: set[str] = set()
MAX_HASHES = 10


def get_download_link() -> str | None:
    try:
        site = requests.get("https://tci72.ru/students/schedule/").text
        soup = BeautifulSoup(site, "html.parser")
        link_element = soup.find("li", class_="file-download").find("a")
        link = link_element.get("href")
        full_link = "https://tci72.ru" + link
        return full_link
    except Exception:
        log.error(f"Failed to parse download link", exc_info=True)
        return None


def start_listing():
    log.info("Worker started, monitoring schedule updates from https://tci72.ru...")
    while True:
        try:
            if len(last_hashes) > MAX_HASHES:
                last_hashes.pop()

            link = get_download_link()
            if not link:
                log.error("Failed to find download link on the website https://tci72.ru...")
                time.sleep(60)
                continue

            response = requests.get(link)
            if response.status_code != 200:
                log.error(f"Failed to download file. Status code: {response.status_code}")
                time.sleep(60)
                continue

            hash_md5 = hashlib.md5(response.content).hexdigest()
            if hash_md5 in last_hashes:
                time.sleep(60)
                continue

            try:
                result = add_schedule_from_xlsx(response.content)
                if result:
                    log.info(f"New schedule added successfully: {link}")
                else:
                    log.info(f"Schedule from site is not added: {link}")
            except Exception:
                log.error(f"Error when processing given schedule {link}.", exc_info=True)
            last_hashes.add(hash_md5)

        except Exception:
            log.error(f"Unexpected error in worker loop", exc_info=True)

        time.sleep(60)


if __name__ == "__main__":
    start_listing()
