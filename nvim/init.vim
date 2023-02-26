call plug#begin("~/.config/nvim/plugged")
 " Better language packs
 Plug 'sheerun/vim-polyglot'
 " Relative numbers (current ln is 0)
 Plug 'ericbn/vim-relativize'
 " Nice icons in the file explorer and file type status line.
 Plug 'ryanoasis/vim-devicons'
 " airline alternative
 Plug 'nvim-lualine/lualine.nvim'
 " If you want to have icons in your statusline choose one of these
 Plug 'kyazdani42/nvim-web-devicons'
 " dependencies
 Plug 'nvim-lua/plenary.nvim'
 Plug 'nvim-lua/popup.nvim'
 " recommended - LSP config
 Plug 'neovim/nvim-lspconfig'
 " diff view
 Plug 'sindrets/diffview.nvim'
 " Conquer of Completion
 Plug 'neoclide/coc.nvim', {'branch': 'release'}
 " Rust
 Plug 'rust-lang/rust.vim'
 " Ansible
 Plug 'pearofducks/ansible-vim'
 " Nerdtree 
 Plug 'preservim/nerdtree'
 " Fuzzy finder
 Plug 'junegunn/fzf', { 'do': { -> fzf#install() } }
 Plug 'junegunn/fzf.vim'
call plug#end()


" tabs and spaces handling
set expandtab
set tabstop=4
set softtabstop=4
set shiftwidth=4

" show line numbers
set relativenumber
set number

" remove ugly vertical lines on window division
set fillchars+=vert:\ 

" utf-8 everywhere
set encoding=utf-8
"
" when scrolling, keep cursor 3 lines away from screen border
set scrolloff=3

" keep BASH as our shell for compatibility
set shell=/bin/bash



" start lualine
lua << END
require('lualine').setup {
    options = {
        theme = 'auto'
        }
}
END

" VimTeX Settings
let g:tex_flavor='latex'
let g:vimtex_quickfix_mode=0
set conceallevel=1
let g:tex_conceal='abdmg'

" ansible
let g:coc_filetype_map = {
  \ 'yaml.ansible': 'ansible',
  \ }

" Nerdtree
nnoremap <C-n> :NERDTree<CR>

" disable mouse
set mouse=
