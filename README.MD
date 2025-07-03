# BDO Profile Scraper
A simple, terminal-based Go application to fetch and save a player's profile data from the Black Desert Online website. This project serves as a Proof of Concept for scraping and parsing the BDO Adventure Log.
The application takes a family name as input, scrapes the corresponding profile page, and saves the structured data as a ProfileData.json file in the same directory.
## Features
Search for any BDO family name.
Parses Family Info, Life Skills, and Character List.
Saves the complete profile to a clean ProfileData.json file.
Built as a Terminal User Interface (TUI) using Go and Bubbletea.
## Prerequisites
To build and run this project from the source code, you will need:
Go (version 1.21 or newer)
Git
Docker (for the containerized build method)

---

## How to Run
There are two primary ways to run this application: building from source or using the provided Dockerfile.
 - **Method 1:** Build from Source (Recommended):\
 This is the simplest way to get started if you have Go installed on your system.

```bash
git clone https://github.com/fernandocarminatti/bdo-tui.git
cd bdo-tui

# Build the application.
go build .

# Linux Run
./bdo-tui

# Windows run (not even tested)
.\bdo-tui.exe
```
The TUI will launch. Enter a family name and press Enter. If successful, a ProfileData.json file will be created in the project directory.\
**Note:** When running with Docker, the ProfileData.json file will be created inside the container and will be deleted when the container exits. To save the file to your local machine, you need to mount a volume:

---

- **Method 2:** Build and Run with Docker:\
This method is ideal if you prefer not to install Go on your local machine. It builds and runs the application inside a clean, isolated Docker container.

```bash
git clone https://github.com/YOUR_USERNAME/bdo-tui.git
cd bdo-tui

# Build image:
docker build -t bdo-scraper .

# Run image:
docker run -it --rm bdo-scraper
```

**Note:** When running with Docker, the ProfileData.json file will be created inside the container and will be deleted when the container exits. To save the file to your local machine, you need to mount a volume:

```bash
# This command maps your current directory (.) to the /app directory inside the container
docker run -it --rm -v "$(pwd):/app" bdo-scraper
```