# 🗂️ makit

A CLI tool that allows you to create files and directories with a single command  
![MIT License](https://img.shields.io/badge/license-MIT-blue "MIT License")
[![Go Report Card](https://goreportcard.com/badge/github.com/kai20020918/makit)](https://goreportcard.com/report/github.com/kai20020918/makit)
[![Coverage Status](https://coveralls.io/repos/github/kai20020918/makit/badge.svg?branch=main)](https://coveralls.io/github/kai20020918/makit?branch=main)

## 👀 Overview

This is a simple yet powerful CLI tool that lets you **create files and directories simultaneously** with a single command.  
No more typing `mkdir` followed by `touch`—this utility streamlines your workflow and boosts productivity, especially for developers and script writers

## 🥞 Usage

`makit` is a powerful CLI tool that simplifies file and directory creation. Once installed, you can use it from any directory in your terminal.

makit [OPTION] <FILES|DIRS...>
**Options:**

- `-p` Create parent directories as needed (default)
- `-m <mode>` Set directory permissions (e.g., -m 755)
- `-d <stamp>` Use specific timestamp (e.g., 202504181200)
- `-c` Do not create file if it doesn’t exist
- `-v` Enable verbose output
- `-h` print this message.

## 🍈 Installation

To use `makit` from any directory in your terminal, you need to place the `makit` executable in a directory that is included in your system's `PATH` environment variable.

**Steps:**

1.  **Build the `makit` executable:**
    Navigate to the root directory of your `makit` project (where `go.mod` is located) and build the application:

    ```bash
    go build -o makit cmd/main/makit.go
    ```

    This will create an executable file named `makit` in your current directory.

2.  **Move the executable to a `PATH` directory:**
    Move the `makit` executable to a common system binary directory, such as `/usr/local/bin/`. This usually requires superuser (sudo) privileges:

    ```bash
    sudo mv makit /usr/local/bin/
    ```

3.  **Verify the installation:**
    Open a new terminal window (or run `source ~/.zshrc` if using zsh) and try running `makit` without `./`:
    ```bash
    makit -h
    ```
    If you see the help message, `makit` is successfully installed and ready to use!

## 🐼 About

### Developer Name

    Kairi Miyazaki

### Icon

<img src="file_dir.png" alt="ファイルとディレクトリが半分のアイコン" width="200">

### Origin of Name

    The name makit is a blend of:

      make – representing the action of creating files and directories, inspired by traditional UNIX tools like make, mkdir, and touch.

      it – referring to “it”, the thing you want to create. Whether it’s a file, a folder, or an entire structure—just makit.

### Version

-
