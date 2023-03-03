export LC_CTYPE="en_US.UTF-8"
xset -b &
xset +fp ~/.fonts
for font in /usr/local/share/fonts/*; do
        xset +fp "${font}"
        done
xset fp rehash

exec /usr/local/bin/i3
