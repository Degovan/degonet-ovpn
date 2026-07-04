package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/irsyadulibad/degonet-ovpn/auth-go/cmd"
)

type CCDService struct {
	CCDDir string
}

func (c *CCDService) dir() string {
	if filepath.IsAbs(c.CCDDir) {
		return c.CCDDir
	}
	execPath, _ := os.Executable()
	return filepath.Join(filepath.Dir(execPath), c.CCDDir)
}

func (c *CCDService) Create(user *cmd.User) (bool, error) {
	dir := c.dir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, fmt.Errorf("create ccd dir: %w", err)
	}

	filePath := filepath.Join(dir, user.Username)
	if _, err := os.Stat(filePath); err == nil {
		return false, nil
	}

	content := fmt.Sprintf("ifconfig-push %s %s\n", user.IP, user.Netmask)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return false, fmt.Errorf("write ccd: %w", err)
	}
	return true, nil
}

func (c *CCDService) Delete(username string) (bool, error) {
	filePath := filepath.Join(c.dir(), username)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, nil
	}
	if err := os.Remove(filePath); err != nil {
		return false, fmt.Errorf("delete ccd: %w", err)
	}
	return true, nil
}
