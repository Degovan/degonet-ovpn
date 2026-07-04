package main

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/irsyadulibad/degonet-ovpn/auth-go/cmd"
)

type AuthService struct{}

func (a *AuthService) Login(username, password string) (*cmd.User, bool, error) {
	user, err := a.FindByUsername(username)
	if err != nil {
		return nil, false, err
	}
	if user == nil {
		return nil, false, nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return user, false, nil
	}
	return user, true, nil
}

func (a *AuthService) AddUser(username, password, ip, netmask string) (*cmd.User, error) {
	existing, err := a.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, nil
	}

	var ipCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM mikrotik_users WHERE ip = ?", ip).Scan(&ipCount); err != nil {
		return nil, err
	}
	if ipCount > 0 {
		return nil, nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	res, err := db.Exec(
		"INSERT INTO mikrotik_users (username, password, ip, netmask) VALUES (?, ?, ?, ?)",
		username, string(hashed), ip, netmask,
	)
	if err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()
	return &cmd.User{
		ID:       int(id),
		Username: username,
		IP:       ip,
		Netmask:  netmask,
	}, nil
}

func (a *AuthService) FindByUsername(username string) (*cmd.User, error) {
	row := db.QueryRow(
		"SELECT id, username, password, ip, netmask FROM mikrotik_users WHERE username = ?",
		username,
	)
	var u cmd.User
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.IP, &u.Netmask)
	if err != nil {
		return nil, nil
	}
	return &u, nil
}

func (a *AuthService) FindByIP(ip string) (*cmd.User, error) {
	row := db.QueryRow(
		"SELECT id, username, password, ip, netmask FROM mikrotik_users WHERE ip = ?",
		ip,
	)
	var u cmd.User
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.IP, &u.Netmask)
	if err != nil {
		return nil, nil
	}
	return &u, nil
}

func (a *AuthService) ListUsers() ([]cmd.User, error) {
	rows, err := db.Query("SELECT id, username, ip, netmask FROM mikrotik_users ORDER BY username ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []cmd.User
	for rows.Next() {
		var u cmd.User
		if err := rows.Scan(&u.ID, &u.Username, &u.IP, &u.Netmask); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (a *AuthService) DeleteUser(username string) (bool, error) {
	res, err := db.Exec("DELETE FROM mikrotik_users WHERE username = ?", username)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
