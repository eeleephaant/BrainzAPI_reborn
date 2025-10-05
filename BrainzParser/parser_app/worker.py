import hashlib
import logging
import time

from bs4 import BeautifulSoup
import requests

from parser_app.ttsiigh_utils import add_schedule_from_xlsx

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)

last_hashes: set[str] = set()
MAX_HASHES = 10


def get_download_link() -> str:
    try:
        site = requests.get("https://tci72.ru/students/schedule/").text
        soup = BeautifulSoup(site, "html.parser")
        link_element = soup.find("li", class_="file-download").find("a")
        link = link_element.get("href")
        full_link = "https://tci72.ru" + link
        return full_link
    except Exception as ex:
        logging.error(f"Failed to parse download link: {ex}")
        return ""


def start_listing():
    logging.info("Worker started, monitoring schedule updates...")
    while True:
        try:
            if len(last_hashes) > MAX_HASHES:
                last_hashes.pop()

            link = get_download_link()
            if not link:
                time.sleep(60)
                continue

            response = requests.get(link)
            if response.status_code != 200:
                logging.warning(f"Failed to download file. Status code: {response.status_code}")
                time.sleep(60)
                continue

            hash_md5 = hashlib.md5(response.content).hexdigest()
            if hash_md5 in last_hashes:
                logging.info("No new schedule found.")
                time.sleep(60)
                continue

            try:
                result = add_schedule_from_xlsx(response.content)
                if result:
                    logging.info(f"[V] New schedule added successfully: {link}")
                else:
                    logging.info(f"[X] Schedule from site is not added: {link}")
            except Exception as ex:
                logging.error(f"Error processing schedule: {ex}")
            last_hashes.add(hash_md5)

        except Exception as ex:
            logging.error(f"Unexpected error in worker loop: {ex}")

        time.sleep(60)


if __name__ == "__main__":
    start_listing()