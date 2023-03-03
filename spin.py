#!/usr/bin/env python3
import time
from sys import argv, exit
from yaspin import yaspin

def sleep():
    msg = "Press Ctrl-C to cancel installation of %s... " % argv[1]
    with yaspin(text=msg):
        time.sleep(3)

if __name__ == '__main__':
    try:
        sleep()
        print("Proceeding...")
    except KeyboardInterrupt:
        print("Cancelled.")
        exit(1)

