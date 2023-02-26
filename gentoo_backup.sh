#!/usr/bin/env bash
# This is supposed to run as a cron job.

running_gentoo=$(cat /etc/os-release | grep -o "ID=gentoo")

if [[ -z ${running_gentoo} ]];then
    printf "Not running Gentoo.Exiting..."
    exit 1
fi

export DOTFILES_DIR=$HOME/dev/dotfiles

cat /var/lib/portage/world > /tmp/gentoo_package_list

new_packages=$(diff /tmp/gentoo_package_list $DOTFILES_DIR/gentoo_package_list)

if [[ -n ${new_packages} ]];then
    cp /tmp/gentoo_package_list $DOTFILES_DIR/gentoo_package_list
    cd $DOTFILES_DIR
    git add .
    git commit -m "Added latest packages"
    git push
    echo "Added latest packages"
fi

exit 0
