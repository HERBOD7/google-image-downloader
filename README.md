# google-image-downloader

Go application for download, resize, and store images in PostgreSQL DB


## Features

- Download images base on specific search query
- Resize images
- Store images in PostgreSQL using Gorm


## Installation

1. Clone the project

    ```bash
      git clone https://github.com/HERBOD7/google-image-downloader.git
      cd google-image-downloader
    ```
2. Setup Environment Variable: create a `.env` file in the project root with these data using [Google Custom Search API](https://console.cloud.google.com/marketplace/product/google/customsearch.googleapis.com)
    ```bash
      GOOGLE_API_KEY="Your-google-api-key"
      GOOGLE_CUSTOM_SEARCH_ENGINE_ID="Your-google-custom-search-engine-id"
    ```
3. Install Dependencies:
    ```bash
      go mod tidy
   ```
4. Create a PostgreSQL database and user for the application

## Usage/Examples

- The application can be configured via commandline flags:
  - `q` (string) : Search query
  - `max` (int) : Max number of images want to download
  - `host` (string) : DB host
  - `p` (int) : DB port
  - `name` (string) : Database Name
  - `u` (string) : Database User
  - `pass` (string) : Database Password
- To download images, use the example following command:
```bash
    go run . -q="kitten" -max=5 -host="localhost" -p=5432 -name="images" -u="postgres" -pass=""
```

