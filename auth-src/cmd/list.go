package cmd

import (
	"github.com/irsyadulibad/degonet-ovpn/auth-go/output"
)

type ListCommand struct {
	Auth AuthOperations
}

func (c *ListCommand) Run() int {
	output.Header("List OpenVPN Users")

	users, err := c.Auth.ListUsers()
	if err != nil {
		output.Error("Gagal mengambil daftar user.")
		return 1
	}

	if len(users) == 0 {
		output.Warning("Belum ada user terdaftar.")
		return 0
	}

	rows := make([]output.Row, 0, len(users))
	for _, u := range users {
		rows = append(rows, output.Row{u.Username, u.IP, u.Netmask})
	}

	output.TableWithHeaders([]string{"Username", "IP", "Netmask"}, rows)
	output.Success("Selesai menampilkan daftar user.")
	return 0
}
