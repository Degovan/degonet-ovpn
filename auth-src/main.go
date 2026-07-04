package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/irsyadulibad/degonet-ovpn/auth-go/cmd"
	"github.com/irsyadulibad/degonet-ovpn/auth-go/output"
)

func main() {
	if err := loadEnv(); err != nil {
		output.Error(fmt.Sprintf("Load env: %v", err))
		os.Exit(1)
	}

	dbFile := getEnv("DB_FILE", "/etc/openvpn/data/users.sqlite")
	if err := initDatabase(dbFile); err != nil {
		output.Error(fmt.Sprintf("Database error: %v", err))
		os.Exit(1)
	}
	defer db.Close()

	auth := &AuthService{}
	ccd := &CCDService{CCDDir: getEnv("CCD_DIR", "/etc/openvpn/ccds")}
	defaultNetmask := getEnv("DEFAULT_NETMASK", "255.255.255.0")

	username := os.Getenv("username")
	if username != "" {
		runAuthMode(auth, ccd, username)
		return
	}

	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		addCmd := &cmd.AddCommand{Auth: auth, CCD: ccd, Netmask: defaultNetmask}
		os.Exit(addCmd.Run(os.Args[2:]))

	case "delete":
		delCmd := &cmd.DeleteCommand{Auth: auth, CCD: ccd}
		os.Exit(delCmd.Run(os.Args[2:]))

	case "list":
		listCmd := &cmd.ListCommand{Auth: auth}
		os.Exit(listCmd.Run())

	default:
		showUsage()
		os.Exit(1)
	}
}

func runAuthMode(auth *AuthService, ccd *CCDService, username string) {
	password := os.Getenv("password")

	user, found, err := auth.Login(username, password)
	if err != nil || user == nil || !found {
		os.Exit(1)
	}

	ccd.Create(user)
	os.Exit(0)
}

func showUsage() {
	output.Header("OpenVPN Auth CLI")
	output.Error("Command tidak dikenal. Gunakan salah satu: add, delete, list")
	fmt.Println("Contoh:")
	fmt.Println("  auth add <username> <ip> [password] [netmask]")
	fmt.Println("  auth delete <username>")
	fmt.Println("  auth list")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadEnv() error {
	envPath := ".env"
	if execPath, err := os.Executable(); err == nil {
		envPath = filepath.Join(filepath.Dir(execPath), ".env")
	}
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		envPath = ".env"
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return nil
}
