#!/usr/bin/env python3
# -*- coding: utf-8; -*-

# The MIT License (MIT)
#
# Copyright (c) 2021 Álan Crístoffer e Sousa
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

import json
import os
import os.path
import re
import shutil
import subprocess
import sys
import dbkp
from typing import Union, Optional
from pydantic.dataclasses import dataclass


def print_help():
    text = """
    Usage:

    dbkp backup [config_file]
        Backups according to config_file

    dbkp restore [config_file]
        Restores a backup according to config_file

    If config_file is not given, it is assumed to be a file
    named dbkp.json in the current directory.

    Dotfiles are copied on backup/restore, not symlinked. rsync -La is used to
    both backup and restore.

    dbkp.json is formatted as follows:

    [                            # root must be a list
        "~/.vimrc",              # same as {"path":"~/.vimrc", "alias":"vimrc"}
        {
            "path": "~/.config/nvim",  # path of the file to backup/restore
            "alias": "neovim",         # name inside the backup folder
            "only": ["init.vim"],      # if :path is a dir, after copy, removes
            "exclude": ["secrets.vim"],# everything but :only, or removes
                                       # :exclude from the backup. :only
                                       # overwrites :exclude.
            "links": [
                "~/.neovim",              # after restoring, symlinks ~/.neovim
                                          # to :path
                ["init.vim", "~/.vimrc"]  # after restoring, symlinks ~/.vimrc
                                          # to :path/init.vim
            ],
        },
        {
            "backup": "brew leaves",         # execs :backup and saves output to
                                             # :alias inside backup
            "restore": "xargs brew install", # execs cat :alias | :restore
            "alias": "brew.leaves"
        }
    ]
    """
    print(text.strip())


def rm_rf(path):
    if os.path.isfile(path) or os.path.islink(path):
        os.remove(path)
    elif os.path.isdir(path):
        shutil.rmtree(path)


@dataclass
class FileBackup:
    path: str
    alias: Optional[str] = None
    only: Optional[list[str]] = None
    exclude: Optional[list[str]] = None
    links: Optional[list[Union[str, list[str]]]] = None

    def __post_init_post_parse__(self):
        self.path = os.path.abspath(os.path.expanduser(self.path))
        if not self.alias:
            self.alias = re.sub(r"^\.", "", os.path.basename(self.path))

    def do_backup(self):
        print(f"Backing up {self.path}")
        dst = os.path.join(os.getcwd(), "dotfiles", self.alias)
        if os.path.isfile(self.path) or os.path.islink(self.path):
            subprocess.Popen(["rsync", "-La", self.path, dst]).wait()
        elif os.path.isdir(self.path):
            src = os.path.join(self.path, "")
            dst = os.path.join(dst, "")
            subprocess.Popen(["rsync", "-La", src, dst]).wait()
            if self.only:
                self.exclude = set(os.listdir(dst)) - set(self.only)
            if self.exclude:
                for entry in self.exclude:
                    entry = os.path.join(dst, entry)
                    print(f"\tExcluding {entry}")
                    rm_rf(entry)
        else:
            raise ValueError(f"File not found: {self.path}")

    def do_restore(self):
        print(f"Restoring {self.path}")
        src = os.path.join(os.getcwd(), "dotfiles", self.alias)
        if not os.path.exists(src):
            print(f"\tFile does not exist: {src}")
            return
        if os.path.isfile(src) or os.path.islink(src):
            subprocess.Popen(["rsync", "-La", src, self.path]).wait()
        elif os.path.isdir(src):
            src = os.path.join(src, "")
            dst = os.path.join(self.path, "")
            subprocess.Popen(["rsync", "-La", src, dst]).wait()
        if self.links:
            for link in self.links:
                if isinstance(link, str):
                    link = os.path.abspath(os.path.expanduser(link))
                    print(f"\tCreating link: {link}")
                    rm_rf(link)
                    os.symlink(self.path, link)
                elif isinstance(link, list):
                    src = os.path.join(self.path, link[0])
                    dst = os.path.abspath(os.path.expanduser(link[1]))
                    print(f"\tCreating link: {dst}")
                    rm_rf(dst)
                    os.symlink(src, dst)


@dataclass
class CommandBackup:
    backup: str
    restore: str
    alias: str

    def do_backup(self):
        print(f"Backing up with command: {self.backup}")
        out = subprocess.check_output(self.backup, shell=True, text=True)
        dst = os.path.join(os.getcwd(), "dotfiles", self.alias)
        with open(dst, "w+") as dst:
            dst.write(out)

    def do_restore(self):
        print(f"Restoring with command: {self.restore}")
        alias = os.path.join(os.getcwd(), "dotfiles", self.alias)
        if not os.path.exists(alias):
            print(f"\tFile does not exist: {alias}")
            return
        with open(alias, "rb") as alias:
            subprocess.run(self.restore, shell=True, input=alias.read())


def into_object(entry):
    if isinstance(entry, str):
        return FileBackup(path=entry)
    if isinstance(entry, dict):
        if "path" in entry:
            return FileBackup(**entry)
        if "backup" in entry:
            return CommandBackup(**entry)
        raise ValueError(f"Unrecognized entry: {entry}")


def parse_config(config_path):
    if not (os.path.isfile(config_path) or os.path.islink(config_path)):
        raise ValueError(f"Configuration file does not exist: {config_path}")
    with open(config_path, "r") as config_file:
        entries = json.load(config_file)
    entries = list(map(into_object, entries))
    return entries


def main():
    if len(sys.argv) > 3 or len(sys.argv) < 2:
        print_help()
        return
    command = sys.argv[1]
    config_path = (sys.argv[2:] + ["dbkp.json"])[0]
    config_path = os.path.abspath(os.path.expanduser(config_path))
    os.chdir(os.path.dirname(config_path))
    if command not in ["--version", "backup", "restore"]:
        print_help()
        return
    if command == "--version":
        print(dbkp.__version__)
        return
    try:
        config = parse_config(config_path)
        print(f"Backup folder: {os.getcwd()}")
        print(f"Configuration file: {config_path}")
        print("")
        dotfiles_folder = os.path.join(os.getcwd(), "dotfiles")
        if command == "backup":
            if os.path.exists(dotfiles_folder):
                shutil.rmtree(dotfiles_folder)
            os.mkdir(dotfiles_folder)
            for cfg in config:
                cfg.do_backup()
        elif command == "restore":
            for cfg in config:
                cfg.do_restore()
    except ValueError as error:
        print(error)


if __name__ == "__main__":
    main()
