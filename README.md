# Sota (Simple One Time Authenticator)

[![GitHub release](https://img.shields.io/github/v/release/febras/sota?color=blue&style=flat)](https://github.com/febras/sota/releases)
[![contributions](https://img.shields.io/badge/contributions-welcome-blue.svg?style=flat)](https://github.com/febras/sota/issues)

<img src="https://raw.githubusercontent.com/FebraS/sota/refs/heads/main/assets/Sota.png" alt="Sota">

Sota is a high-performance command-line utility built in Go for managing and decoding multi-account migration QR codes.

It is built using Go for fast execution, but it relies on a Python script under the hood to actually scan the QR code. We use this hybrid approach because standard Go libraries often struggle to read highly dense QR codes, while Python's `pyzbar` library can read them perfectly.

## Features

*   **Hybrid Engine**: Uses Go for the main application and CLI speed, while delegating the heavy lifting of image scanning to Python.
*   **High Accuracy**: Easily scans and decodes very dense migration QR codes that usually cause errors in standard scanner tools.
*   **Direct Export**: Automatically extracts the `otpauth-migration://` URIs and saves them directly to a text file for easy backup.
*   **Cross Platform**: Works on Windows, macOS, and Linux as long as Python is installed.

## Requirements

Before using Sota, you need to have the following installed on your computer:
1.  **Go**: To build the application.
2.  **Python 3**: To run the background scanning script.

You also need to install two Python libraries. Choose the installation method based on your Operating System:

#### Option A: Linux Package Manager (Recommended for Debian/Ubuntu/Kali/Arch)
Installing dependencies directly via your Linux package manager ensures that all underlying C-libraries (like `libzbar`) are properly configured and avoids system-managed environment restrictions:

* **Debian / Ubuntu / Kali Linux**:

```bash
sudo apt update && sudo apt install -y python3-pyzbar python3-pillow
```

* **Arch**:
```bash
sudo pacman -S python-pyzbar python-pillow
```

* **Fedora / RHEL**:
```bash
sudo dnf install python3-pyzbar python3-pillow
```

#### Option B: Via PIP (Windows, macOS, or Linux venv)
If you prefer using pip (or are using a Python virtual environment):

Install Python libraries:

```bash
pip install pyzbar pillow
```
Required for PIP users on Linux: You must manually install the ZBar C-library shared package for pyzbar to function:

* **Debian/Ubuntu/Kali:** 
```bash
sudo apt install libzbar0
```
* **Arch Linux:**
```bash
sudo pacman -S zbar
```

## Setup and Installation

### Option 1: Download Pre-compiled Binary (Fastest)
You don't need to build from source! Download the latest ready-to-use binary for your OS from the Releases page:

1. Go to [Sota Releases](https://github.com/febras/sota/releases).
2. Download the appropriate binary for your system (Windows, macOS, or Linux).
3. Place the executable in your working directory.

---

### Option 2: Build from Source

First, clone the repository:
```bash
git clone https://github.com/febras/sota.git
cd sota
```

If you prefer to compile the application yourself:

**Using Make (Recommended)**
You can use the provided Makefile to quickly build or clean the project:

**Build binary for your current OS:**

```bash
make build
```
(Output will be saved in the build directory).


Build binaries for all platforms (Linux, macOS Intel/Apple Silicon, Windows):

```bash
make release-all
```

Clean built files:
```bash
make clean
```

Without Make (Manual Go Build)
If make is not available on your environment, compile directly using Go:

```bash
go build -o bin/sota .
```

### Usage
Run the compiled application by pointing it to your QR code image and specifying the desired output file for the extracted URIs:

```bash
./sota -i qrcode.jpeg -o accounts.txt
```
(On Windows, use .\build\sota-windows-amd64.exe -i qrcode.jpeg -o accounts.txt or point directly to the downloaded executable).

If the extraction is successful, your URIs will be cleanly formatted and saved inside accounts.txt.

## Advanced Features
Sota includes several arguments for control and flexibility during execution.

| Argumen | Description |
|---|---|
|`--import-migration <path_to_image>`	| Scans a Google Authenticator migration QR code and adds all URIs to accounts.txt. |
|`--output-file <filename.txt>`	| Use with --import-migration to save the imported URIs to a custom file. |
|`--generate-ykman <path_to_image>`	| Converts URIs from a QR code into ready-to-run ykman commands for YubiKey. |
|`--export <filename.txt>`	| Generates individual QR code images for each account in the specified file (defaults to accounts.txt). |
|`--read <filename.txt>`	| Loads accounts from a custom file instead of accounts.txt. |
|`--help`	| Displays a brief description of the program and all available arguments. |

## Contribution
Contributions are welcome. If you are interested in improving the codebase, adding new features, or fixing bugs, please feel free to open an issue or submit a pull request on the GitHub repository.

## License
This project is open-source and distributed under the Apache-2.0. 
<br>See the LICENSE file for more details.