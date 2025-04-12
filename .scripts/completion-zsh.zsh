#!/usr/bin/env zsh

tex completion zsh > _tex
sudo mv ./_tex $fpath[1]
