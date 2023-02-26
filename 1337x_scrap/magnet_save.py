#!/usr/bin/env python3
# Scrap 1337x.to and collect magnet links easily

from time import sleep
from py1337x import py1337x

def main():
    parse_torrents()
    save_magnets()
    magnets_to_file()

search_string = input("Enter a search string: ")
page_count = int(input("Enter page count: ") or 10)
proxy_address = str(input("Enter proxy address: ") or "1337x.to")


print('Searching for:',search_string, 'Pages:', page_count, 'Proxy:', proxy_address)
torrents = py1337x(proxy=proxy_address)

current_page_torrents = {}
all_torrent_info = []
all_magnet_links = []
current_torrent = {}


def parse_torrents():
    for i in range(1, page_count+1):
        current_page_torrents = torrents.search(search_string, i)
        for result in current_page_torrents['items']:
            print('Parsing', result['name'])
            # Here we collect the desired torrent links
            all_torrent_info.append(result['link'])

# For each link found, keep its magnet link
def save_magnets():
    for link in all_torrent_info:
        current_torrent = torrents.info(link)
        print('Saving magnet', current_torrent['name'])
        all_magnet_links.append(current_torrent['magnetLink'])

def magnets_to_file():

    if not all_magnet_links:
        print("No torrents found.")
        return None

    # if magnet links have been collected
    target_filename = search_string + ' magnets.txt'
    with open(target_filename, 'w') as f:
        for line in all_magnet_links:
            f.write(f"{line}\n")
    print(len(all_magnet_links), "torrents saved to", target_filename, "\nDone.")


if __name__ == "__main__":
    main()
    exit(0)
