package cmd

import (
	"fmt"

	"github.com/irsyadulibad/degonet-ovpn/auth-go/output"
)

type DeleteCommand struct {
	Auth AuthOperations
	CCD  CCDOperations
}

func (c *DeleteCommand) Run(args []string) int {
	if len(args) < 1 || args[0] == "" {
		output.Error("Usage: auth delete <username>")
		fmt.Println("Contoh: auth delete budi")
		return 1
	}

	username := args[0]

	user, err := c.Auth.FindByUsername(username)
	if err != nil || user == nil {
		output.Error(fmt.Sprintf("User '%s' tidak ditemukan.", username))
		return 1
	}

	deleted, err := c.Auth.DeleteUser(username)
	if err != nil || !deleted {
		output.Error(fmt.Sprintf("Gagal menghapus user '%s' dari database.", username))
		return 1
	}

	ccdDeleted, _ := c.CCD.Delete(username)
	ccdStatus := "dihapus"
	if !ccdDeleted {
		ccdStatus = "tidak ditemukan"
	}

	output.Header("Delete OpenVPN User")
	output.Table([]output.Row{
		{"Username", username},
		{"Status DB", "dihapus"},
		{"Status CCD", ccdStatus},
	})
	output.Success("User berhasil dihapus.")
	return 0
}
