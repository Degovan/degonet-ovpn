package cmd

type AuthOperations interface {
	AddUser(username, password, ip, netmask string) (*User, error)
	FindByUsername(username string) (*User, error)
	FindByIP(ip string) (*User, error)
	DeleteUser(username string) (bool, error)
	ListUsers() ([]User, error)
}

type CCDOperations interface {
	Create(user *User) (bool, error)
	Delete(username string) (bool, error)
}

type User struct {
	ID       int
	Username string
	Password string
	IP       string
	Netmask  string
}
