package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"

	"github.com/gagehenrich/bucket/cli/internal/api"
	"github.com/gagehenrich/bucket/cli/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Config error:", err)
		return
	}

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {

	case "account":
		handleAccount(cfg)
		return
	//case "keys":
	//	handleKeys(cfg)
	//	return
	case "push":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bucket push <file>")
			return
		}
		handlePush(cfg, os.Args[2])
		return
	case "pull":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bucket pull <tiny-url>")
			return
		}
		handlePull(cfg, os.Args[2])
		return
	case "list":
		handleList(cfg)
		return

	default:
		printHelp()
	}
}

func readSecret(prompt string) string {
	fmt.Print(prompt)

	byteSecret, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	return strings.TrimSpace(string(byteSecret))
}

func readPassword() string {
	fmt.Print("Password: ")

	bytePassword, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	return strings.TrimSpace(string(bytePassword))
}


func printHelp() {
	fmt.Println(`bucket CLI — Secure File Sharing

Commands:
  bk account            Login or create an account
  bk push <file>        Upload a file
  bk pull <tiny-url>    Download a file
  bk list               List uploaded files
`)
}

//
// ------------------------------------------------------------
// Account
// ------------------------------------------------------------
//

func handleAccount(cfg *config.Config) {
    // Check if already logged in
    if cfg.APIKey != "" {
        fmt.Println("Already logged in with API key:", cfg.APIKey)
        fmt.Print("Create new API key? (y/n): ")
        reader := bufio.NewReader(os.Stdin)
        answer, _ := reader.ReadString('\n')
        if strings.TrimSpace(strings.ToLower(answer)) != "y" {
            return
        }
    }
    
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Email: ")
    email, _ := reader.ReadString('\n')
    email = strings.TrimSpace(email)

    password := readPassword()

    client := api.New(cfg)

    apiKey, err := client.LoginOrCreate(email, password)
    if err != nil {
        fmt.Println("Account error:", err)
        return
    }

    cfg.APIKey = apiKey
    config.Save(cfg)

    fmt.Println("Account ready.")
    fmt.Println("API Key:", apiKey)
}

//
// ------------------------------------------------------------
// Push (Upload)
// ------------------------------------------------------------
//

func handlePush(cfg *config.Config, filepath string) {
	if cfg.APIKey == "" {
		fmt.Println("Not logged in. Run: bucket account")
		return
	}

	stat, err := os.Stat(filepath)
	if err != nil {
		fmt.Println("File error:", err)
		return
	}

	client := api.New(cfg)

	uploadInit, err := client.RequestUpload(stat.Name(), stat.Size())
	if err != nil {
		fmt.Println("Upload failed:", err)
		return
	}

	err = client.UploadFile(uploadInit.UploadURL, filepath)
	if err != nil {
		fmt.Println("Upload failed:", err)
		return
	}

	fmt.Println("Upload complete!")
	fmt.Println("Tiny URL:", cfg.APIBase+"/d/"+uploadInit.TinyCode)
	fmt.Println("Download secret:", uploadInit.Secret)
	fmt.Println("Expires:", uploadInit.ExpiresAt)
}

//
// ------------------------------------------------------------
// Pull (Download)
// ------------------------------------------------------------
//

func handlePull(cfg *config.Config, tinyURL string) {
    tiny := api.ExtractTinyCode(tinyURL)
    secret := readSecret("Enter secret: ")

    client := api.New(cfg)

    // authenticate presigned URL
    downloadURL, filename, err := client.AuthDownload(tiny, secret)  
    if err != nil {
        fmt.Println("Download auth failed:", err)
        return
    }

	// fmt.Println("DEBUG: filename from server =", filename)

    // download object
    savedFilename, err := client.DownloadFile(downloadURL, filename) 
    if err != nil {
        fmt.Println("Download failed:", err)
        return
    }

    fmt.Println("Downloaded:", savedFilename)
}

//
// ------------------------------------------------------------
// List Files
// ------------------------------------------------------------
//

func handleList(cfg *config.Config) {
	client := api.New(cfg)

	files, err := client.ListFiles()
	if err != nil {
		fmt.Println("List failed:", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No files.")
		return
	}

	fmt.Printf("%-10s %-20s %-12s %-24s\n", "Tiny", "Filename", "Size", "Expires")
	fmt.Println(strings.Repeat("-", 70))

	for _, f := range files {
		fmt.Printf("%-10s %-20s %-12s %-24s\n",
			f.TinyCode,
			f.Filename,
			humanSize(f.SizeBytes),
			f.ExpiresAt,
		)
	}
}

//
// ------------------------------------------------------------
// Human-readable size formatting
// ------------------------------------------------------------
//

func humanSize(n int64) string {
	if n < 1024 {
		return strconv.FormatInt(n, 10)
	}
	kb := float64(n) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := kb / 1024
	if mb < 1024 {
		return fmt.Sprintf("%.1f MB", mb)
	}
	gb := mb / 1024
	return fmt.Sprintf("%.2f GB", gb)
}
