package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

type Image struct {
	gorm.Model
	Filename string
	Data     []byte
}

const endpoint = "https://www.googleapis.com/customsearch/v1"

func search(query string) ([]string, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	cx := os.Getenv("GOOGLE_CUSTOM_SEARCH_ENGINE_ID")

	encodedQuery := url.QueryEscape(query)

	searchUrl := fmt.Sprintf("%s?q=%s&key=%s&cx=%s&searchType=image", endpoint, encodedQuery, apiKey, cx)

	res, err := http.Get(searchUrl)
	if err != nil {
		fmt.Println("1", err)
		return nil, err
	}

	defer res.Body.Close()

	var result struct {
		Items []struct {
			Link string `json:"link"`
		} `json:"items"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		fmt.Println("2", err)
		return nil, err
	}

	var urls []string
	for _, item := range result.Items {
		urls = append(urls, item.Link)
	}

	return urls, nil
}

func downloadImg(url, name string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	return err
}

func connectDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func storeImage(db *gorm.DB, imagePath string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}

	defer file.Close()

	imgData, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	img := Image{
		Filename: file.Name(),
		Data:     imgData,
	}

	result := db.Create(&img)
	return result.Error
}

func resizeImage(filePath string, width, height int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	img, format, err := image.Decode(file)
	if err != nil {
		file.Close()
		return err
	}
	file.Close()

	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.ApproxBiLinear.Scale(newImg, newImg.Bounds(), img, img.Bounds(), draw.Over, nil)

	outPutFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outPutFile.Close()

	switch strings.ToLower(format) {
	case "jpeg":
		return jpeg.Encode(outPutFile, newImg, nil)
	case "png":
		return png.Encode(outPutFile, newImg)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func main() {
	var (
		query      string
		maxImg     int
		host       string
		port       int
		dbName     string
		dbUser     string
		dbPassword string
	)

	flag.StringVar(&query, "q", "", "Search query for images")
	flag.IntVar(&maxImg, "max", 10, "Max number of images want to download")
	flag.StringVar(&host, "host", "localhost", "DB host")
	flag.IntVar(&port, "p", 5432, "DB port")
	flag.StringVar(&dbName, "name", "images", "Database Name")
	flag.StringVar(&dbUser, "u", "postgres", "Database User")
	flag.StringVar(&dbPassword, "pass", "", "Database Password")

	flag.Parse()

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", host, port, dbUser, dbName, dbPassword)
	db, err := connectDB(dsn)
	if err != nil {
		fmt.Printf("Failed to connect to DB: %v\n", err)
		return
	}

	err = db.AutoMigrate(&Image{})
	if err != nil {
		return
	}

	urls, err := search(query)
	if err != nil {
		fmt.Println("search failed:", err)
		return
	}

	if _, err := os.Stat("images"); os.IsNotExist(err) {
		os.Mkdir("images", 0755)
	}

	for i := 0; i < maxImg; i++ {
		url := urls[i]
		fileName := filepath.Join("images", fmt.Sprintf("%s-%d.jpg", query, i))
		if err := downloadImg(url, fileName); err != nil {
			fmt.Printf("Failed to download %s: %s\n", url, err)
		} else {
			fmt.Println("Downloaded:", fileName)
			imagePath := fileName
			if err := resizeImage(fileName, 800, 600); err != nil {
				fmt.Printf("Failed resizing image: %v\n", err)
				return
			}

			if err := storeImage(db, imagePath); err != nil {
				fmt.Printf("Failed to store images to DB: %v\n", err)
				return
			}
			fmt.Println("Images stored successfully")
		}
	}

}
