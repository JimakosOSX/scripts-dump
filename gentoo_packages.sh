#!/usr/bin/env bash

# exit on error
set -e 

# Based on /var/lib/portage/world
readarray -t packages < gentoo_package_list

emerge --noreplace -a -t --unordered-display --quiet-build ${packages[@]}

echo Enjoy!
