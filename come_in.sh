#!/bin/bash

set -e 

# copy neovim configuration
mkdir -pv ~/.config/nvim
cp -v init.vim ~/.config/nvim/init.vim

# Install vim-plug for neovim
sh -c 'curl -fLo "${XDG_DATA_HOME:-$HOME/.local/share}"/nvim/site/autoload/plug.vim --create-dirs \
       https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim'

# Read packages for neovim
readarray -t packages < nvim_pkg_list

# Install packages based on package manager
if [[ -x /usr/bin/apt ]];then
    sudo apt install -y ${packages[@]}

elif [[ -x /usr/bin/pacman ]];then
    sudo pacman -S ${packages[@]}

elif [[ -x /usr/bin/dnf ]];then
    sudo dnf install -y ${packages[@]}
fi

# Set nvim as default in bash
cat >> ~/.bashrc << EOF
export EDITOR=nvim
alias vi=$EDITOR
alias vim=$EDITOR
EOF

# ... and in zsh
cat >> ~/.zshrc << EOF
export EDITOR=nvim
alias vi=$EDITOR
alias vim=$EDITOR
EOF

echo "nvim +PlugInstall +qall"

# Bad way of getting our fonts
curl -s https://api.github.com/repos/ryanoasis/nerd-fonts/releases/latest | grep "browser_download_url.*Meslo.zip" | cut -d : -f2,3 | tr -d \"| wget -qi -

mkdir -pv ~/.fonts
mv Meslo.zip ~/.fonts
cd ~/.fonts
unzip Meslo.zip
rm Meslo.zip
cd -

echo "Neovim installed and ready."
exit 0
