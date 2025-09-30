import traceback
from time import sleep

import ttsiigh_utils

def start():
    print("Brainz Parser v1")
    while True:
        try:
            ttsiigh_utils.refresh_links()
            print("---- Loaded")
        except Exception as e:
            print(f"error: {e}")
            print(f"{traceback.format_exc()}")

        sleep(30)


if __name__ == "__main__":
    start()
