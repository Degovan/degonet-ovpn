package cmd

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/irsyadulibad/degonet-ovpn/auth-go/output"
)

type AddCommand struct {
	Auth     AuthOperations
	CCD      CCDOperations
	Netmask  string
}

func (c *AddCommand) Run(args []string) int {
	if len(args) < 2 {
		output.Error("Usage: auth add <username> <ip> [password] [netmask]")
		fmt.Println("Contoh: auth add budi 10.8.0.10 rahasia 255.255.255.0")
		return 1
	}

	username := args[0]
	ip := args[1]
	password := ""
	if len(args) > 2 {
		password = args[2]
	}
	netmask := c.Netmask
	if len(args) > 3 {
		netmask = args[3]
	}

	if errs := validateAdd(username, ip, netmask); len(errs) > 0 {
		for _, e := range errs {
			output.Error(e)
		}
		return 1
	}

	if password == "" {
		password = username
	}

	existingUser, _ := c.Auth.FindByUsername(username)
	if existingUser != nil {
		output.Error(fmt.Sprintf("User '%s' sudah ada.", username))
		return 1
	}

	existingIP, _ := c.Auth.FindByIP(ip)
	if existingIP != nil {
		output.Error(fmt.Sprintf("IP '%s' sudah dipakai oleh user '%s'.", ip, existingIP.Username))
		return 1
	}

	user, err := c.Auth.AddUser(username, password, ip, netmask)
	if err != nil {
		output.Error("Gagal menambahkan user. Pastikan username dan IP belum terpakai.")
		return 1
	}

	ccdCreated, _ := c.CCD.Create(user)
	ccdStatus := "dibuat"
	if !ccdCreated {
		ccdStatus = "sudah ada (skip)"
	}

	output.Header("Add OpenVPN User")
	output.Table([]output.Row{
		{"Username", user.Username},
		{"IP", user.IP},
		{"Netmask", user.Netmask},
		{"Password", "****"},
		{"Status CCD", ccdStatus},
	})
	output.Success("User berhasil ditambahkan.")

	if !ccdCreated {
		output.Warning("User DB berhasil ditambah, tapi file CCD sudah ada sehingga tidak dibuat ulang.")
	}
	return 0
}

func validateAdd(username, ip, netmask string) []string {
	var errs []string

	if username == "" || ip == "" {
		return []string{fmt.Sprintf("Usage: %s add <username> <ip> [password] [netmask]", os.Args[0])}
	}

	if !isValidIPv4(ip) {
		errs = append(errs, "IP tidak valid. Gunakan format IPv4, contoh: 10.8.0.10")
	}
	if !isValidIPv4(netmask) {
		errs = append(errs, "Netmask tidak valid. Gunakan format IPv4, contoh: 255.255.255.0")
	}
	return errs
}

func isValidIPv4(value string) bool {
	ip := net.ParseIP(strings.TrimSpace(value))
	return ip != nil && ip.To4() != nil
}
