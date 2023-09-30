# Sessionerr

**THIS SCRIPT IS NOT WRITTEN BY ME**
All credit goes to @KyleSanderson
Original repo: https://github.com/KyleSanderson/sessionerr

## What is Sessionerr?

Sessionerr is a Go program designed to specifically tackle the problem mentioned here;
https://github.com/cross-seed/cross-seed/issues/365
It perform a specific task with qBittorrent and Cross-Seed. It checks for completed torrents in qBittorrent, exports them, and submits them to Cross-Seed for cross-seeding.

## Running Breakdown

1. It reads environment variables for configuration.
2. It creates a new session with qBittorrent.
3. It retrieves the list of torrents from qBittorrent.
4. It processes each torrent to check if it's seeding and if it's in a specific save path.
5. If a torrent meets the criteria, it attempts to add it to Cross-Seed.
6. It logs the result of attempting to add each torrent to Cross-Seed.

## How to Build and Run

### Prerequisites

Before you begin, make sure you have the following prerequisites installed:

- Docker: [Install Docker](https://docs.docker.com/get-docker/)
- Docker Compose (usually comes with Docker): [Install Docker Compose](https://docs.docker.com/compose/install/)

### Build and Run with Docker Compose

1. Clone this repository to your local machine:

   ```bash
   git clone https://github.com/yourusername/sessionerr.git
   cd sessionerr

2. Modify the docker-compose.yml file to set the environment

3. Build and start the Docker container:

bash
Copy code
docker-compose up --build
