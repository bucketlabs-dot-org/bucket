package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	
	"golang.org/x/term"

	"github.com/google/uuid"

	"github.com/bucketlabs-dot-org/bucket/cli/internal/api"
	"github.com/bucketlabs-dot-org/bucket/cli/internal/config"
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
	case "del":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bucket del <id>")
			return
		}
		handleDelete(cfg, os.Args[2])
		return
	case "login":
		handleLogin(cfg)
		return
	case "logout": 
		handleLogout(cfg)
		return 
	case "push":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bucket push <file>")
			return
		}
		handlePush(cfg, os.Args[2])
		return
	case "pull":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bucket pull <bURL>")
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

//
// ------------------------------------------------------------
//  LOGIN/LOGOUT
// ------------------------------------------------------------
//
func handleLogout(cfg *config.Config) {
	if cfg.APIKey == "" {
		fmt.Println("You are not currently logged in.")
		return
	}

	client := api.New(cfg)

	// Best-effort server-side logout
	_ = client.Logout()

	deleteJson()
}


func handleLogin(cfg *config.Config) {
	if cfg.APIKey != "" {
		fmt.Println("Already logged in with API key:", cfg.APIKey)
		fmt.Print("Log out? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')

		if strings.TrimSpace(strings.ToLower(answer)) == "y" {
			deleteJson()
		}
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	password := readPassword()

	ensureDeviceIdentity(cfg, email)

	// client now has device_id + name
	client := api.New(cfg)

	// --- 2FA FLOW ---
	var apiKey string
	var err error
	var otpCode string

	apiKey, err = client.Login(email, password, otpCode)

	if err != nil {
		if _, ok := err.(*api.TwoFARequiredError); ok {
			fmt.Printf("2FA code has been sent to %s", email)
			otpCode = readSecret("\n2FA code: ")
			fmt.Println("Retrying with 2FA code...")
			apiKey, err = client.Login(email, password, otpCode)
		}
	}

	if err != nil {
		fmt.Println("Account error:", err)
		return
	}

	cfg.APIKey = apiKey
	_ = config.Save(cfg)

	fmt.Println("Account ready.")
	fmt.Println("API Key:", apiKey)
}


//
// ------------------------------------------------------------
//  ACCOUNT 
// ------------------------------------------------------------
//
func handleAccount(cfg *config.Config) {
	if cfg.APIKey == "" {
		fmt.Println("Not logged in. Run: bucket login")
		return
	}

	client := api.New(cfg)

	info, err := client.FetchAccountInfo()
	if err != nil {
		fmt.Println("Error: Fetch failed:", err)
	} else {
		cfg.Tier = info.Tier
		cfg.UsedBytes = info.UsedBytes
		cfg.Quota = info.Quota
		_ = config.Save(cfg)
	}

	fmt.Println("Account Info")
	fmt.Println("------------")
	fmt.Println("Subscription:", cfg.Tier)
	fmt.Println("Used:", humanSize(cfg.UsedBytes))
	fmt.Println("Quota:", humanSize(cfg.Quota))
	fmt.Println()

	if cfg.Tier != "premium" && cfg.Tier != "bkt_dev" {
		fmt.Println("To increase storage limits, visit:")
		fmt.Println("  https://bucketlabs.org/auth")
	}
}

//
// ------------------------------------------------------------
//  PUSH
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

    // Setup interrupt handler for cleanup
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    uploadDone := make(chan error, 1)
    spinnerDone := make(chan bool)
    fileID := uploadInit.FileID

    // Cleanup function
    cleanup := func() {
        fmt.Println("\n\n⚠️  Upload interrupted. Cleaning up...")
        if err := client.CleanupFailedUpload(fileID); err != nil {
            fmt.Println("Warning: Failed to cleanup incomplete upload:", err)
        } else {
            fmt.Println("✓ Incomplete upload removed")
        }
    }

    // Start spinner
    go showSpinner(spinnerDone, "Uploading...")

    // Upload in goroutine
    go func() {
        uploadDone <- client.UploadFile(uploadInit.UploadURL, filepath)
    }()

    // Wait for upload or interrupt
    select {
    case err := <-uploadDone:
        spinnerDone <- true // Stop spinner
        signal.Stop(sigChan)
        if err != nil {
            fmt.Println("Upload failed:", err)
            cleanup()
            return
        }
    case <-sigChan:
        spinnerDone <- true // Stop spinner
        cleanup()
        os.Exit(130) // Standard exit code for SIGINT
    }

    // VERIFY UPLOAD SUCCESS
    fmt.Print("⏳ Verifying upload...")
    err = client.VerifyUpload(uploadInit.FileID)
    if err != nil {
        fmt.Println("\n❌ Upload verification failed:", err)
        fmt.Println("The file was not pushed. Please try again")
        cleanup()
        return
    }

    fmt.Println("\n\n\t   ✓ Upload complete!\n")
    fmt.Println("    bID: ", uploadInit.TinyCode)
    fmt.Println("   bURL: ", "api.bucketlabs.org"+"/d/"+uploadInit.TinyCode)
    fmt.Println(" Secret: ", uploadInit.Secret)
    fmt.Println("Expires: ", uploadInit.ExpiresAt)
}

//
// ------------------------------------------------------------
//  DELETE
// ------------------------------------------------------------
//
func handleDelete(cfg *config.Config, tiny string) {
    client := api.New(cfg)

    if err := client.DeleteFile(tiny); err != nil {
		if strings.Contains(err.Error(), "invalid subscription") {
			fmt.Println("Error: Unauthorized.\nTo manage your subscription, visit: https://bucketlabs.org/auth")
		} else {
			fmt.Println("Error:", err)
		}
        return
    }

    fmt.Println("Deleted:", tiny)
}

//
// ------------------------------------------------------------
//  PULL
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

    // Start spinner
    spinnerDone := make(chan bool)
    downloadDone := make(chan error, 1)
    
    go showSpinner(spinnerDone, "Downloading")

    // download object
    go func() {
        _, err := client.DownloadFile(downloadURL, filename)
        downloadDone <- err
    }()

    // Wait for download
    err = <-downloadDone
    spinnerDone <- true

    if err != nil {
        fmt.Println("Download failed:", err)
        return
    }

    fmt.Println("\n✓ Downloaded:", filename)
}

//
// ------------------------------------------------------------
//  LIST
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
		fmt.Println("No files in your bucket.")
		return
	}

	fmt.Printf("%-16s %-20s %-12s %-24s\n", "ID", "Filename", "Size", "Expires")
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
//  HELPER FUNCS
// ------------------------------------------------------------
//
func generateDeviceName(email string) string {
	prefix := strings.Split(email, "@")[0]
	prefix = strings.Split(prefix, ".")[0]

	if len(prefix) > 7 {
		prefix = prefix[:7]
	}

	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(90000) + 10000 // 5 digits

	osName := runtime.GOOS // linux, darwin, windows

	return fmt.Sprintf("%s-%d-%s", prefix, num, osName)
}

func ensureDeviceIdentity(cfg *config.Config, email string) {
	changed := false

	if cfg.DeviceID == "" {
		cfg.DeviceID = uuid.NewString()
		changed = true
	}

	if cfg.DeviceName == "" {
		cfg.DeviceName = generateDeviceName(email)
		changed = true
	}

	if changed {
		_ = config.Save(cfg)
	}
}

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

func deleteJson() {
	// DELETE CONFIG
	configFile := config.Path() 
	os.Remove(configFile)
	fmt.Println("Logged out. API key cleared.")
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

func showSpinner(done chan bool, message string) {
    spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
    i := 0
    for {
        select {
        case <-done:
            fmt.Print("\r \r") // Clear the spinner line
            return
        default:
            fmt.Printf("\r%s %s", spinners[i%len(spinners)], message)
            i++
            time.Sleep(100 * time.Millisecond)
        }
    }
}

func printHelp() {
	fmt.Println(`bucket CLI - Secure File Sharing                    
(c) Bucket Labs 2025 

Commands:
  bucket login   	       	Login 
  bucket logout 		Logout 
  bucket account 		View account info
  bucket push <file>        	Upload a file
  bucket pull <bURL>    	Download a file
  bucket list               	List uploaded files
  bucket del <id>		Delete file 

You must first create an account: https://bucketlabs.org/auth`)
}
