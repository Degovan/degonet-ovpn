package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/irsyadulibad/degonet-ovpn/auth-go/output"
)

type DeleteCommand struct {
	Auth AuthOperations
	CCD  CCDOperations
}

func (c *DeleteCommand) Run(args []string) int {
	skipConfirm := false
	var usernames []string
	seen := map[string]bool{}
	for _, a := range args {
		switch a {
		case "-y", "--yes":
			skipConfirm = true
		case "":
		default:
			if !seen[a] {
				seen[a] = true
				usernames = append(usernames, a)
			}
		}
	}

	stdinIsTTY := term.IsTerminal(int(os.Stdin.Fd()))

	if len(usernames) == 0 {
		if !stdinIsTTY {
			output.Error("Usage: auth delete <username> [username...] [-y]")
			fmt.Println("Contoh: auth delete budi ani")
			fmt.Println("Jalankan 'auth delete' tanpa argumen di terminal untuk memilih user secara interaktif.")
			return 1
		}

		picked, code := c.pickUsers()
		if picked == nil {
			return code
		}
		usernames = picked
	}

	// Hapus satu user lewat argumen tetap tanpa konfirmasi seperti sebelumnya.
	needConfirm := stdinIsTTY && !skipConfirm && (len(args) == 0 || len(usernames) > 1)
	if needConfirm && !confirmDelete(usernames) {
		output.Warning("Dibatalkan, tidak ada user yang dihapus.")
		return 1
	}

	return c.deleteAll(usernames)
}

// pickUsers mengembalikan nil jika tidak ada yang dipilih, dengan exit code-nya.
func (c *DeleteCommand) pickUsers() ([]string, int) {
	users, err := c.Auth.ListUsers()
	if err != nil {
		output.Error("Gagal mengambil daftar user.")
		return nil, 1
	}
	if len(users) == 0 {
		output.Warning("Belum ada user terdaftar.")
		return nil, 0
	}

	nameWidth, ipWidth := 0, 0
	for _, u := range users {
		nameWidth = max(nameWidth, len(u.Username))
		ipWidth = max(ipWidth, len(u.IP))
	}
	labels := make([]string, len(users))
	keys := make([]string, len(users))
	for i, u := range users {
		labels[i] = fmt.Sprintf("%-*s  %-*s  %s", nameWidth, u.Username, ipWidth, u.IP, u.Netmask)
		keys[i] = u.Username + " " + u.IP
	}

	idx, err := multiSelect("Pilih user yang akan dihapus", labels, keys)
	if errors.Is(err, errSelectCancelled) {
		output.Warning("Dibatalkan, tidak ada user yang dihapus.")
		return nil, 1
	}
	if err != nil {
		output.Error(fmt.Sprintf("Gagal membuka pilihan interaktif: %v", err))
		return nil, 1
	}
	if len(idx) == 0 {
		output.Warning("Tidak ada user yang dipilih.")
		return nil, 0
	}

	picked := make([]string, len(idx))
	for i, n := range idx {
		picked[i] = users[n].Username
	}
	return picked, 0
}

func confirmDelete(usernames []string) bool {
	output.Header(fmt.Sprintf("Akan menghapus %d user", len(usernames)))
	for _, u := range usernames {
		fmt.Printf("  - %s\n", u)
	}
	fmt.Print("\nLanjutkan? [y/N] ")

	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes" || answer == "ya"
}

func (c *DeleteCommand) deleteAll(usernames []string) int {
	rows := make([]output.Row, 0, len(usernames))
	failed := 0

	for _, username := range usernames {
		dbStatus, ccdStatus := c.deleteOne(username)
		if dbStatus != "dihapus" {
			failed++
		}
		rows = append(rows, output.Row{username, dbStatus, ccdStatus})
	}

	output.Header("Delete OpenVPN User")
	output.TableWithHeaders([]string{"Username", "Status DB", "Status CCD"}, rows)

	deleted := len(usernames) - failed
	if failed > 0 {
		output.Error(fmt.Sprintf("%d user dihapus, %d gagal.", deleted, failed))
		return 1
	}
	output.Success(fmt.Sprintf("%d user berhasil dihapus.", deleted))
	return 0
}

func (c *DeleteCommand) deleteOne(username string) (dbStatus, ccdStatus string) {
	user, err := c.Auth.FindByUsername(username)
	if err != nil || user == nil {
		return "tidak ditemukan", "-"
	}

	deleted, err := c.Auth.DeleteUser(username)
	if err != nil || !deleted {
		return "gagal", "-"
	}

	ccdDeleted, err := c.CCD.Delete(username)
	switch {
	case err != nil:
		return "dihapus", "gagal"
	case !ccdDeleted:
		return "dihapus", "tidak ditemukan"
	}
	return "dihapus", "dihapus"
}
