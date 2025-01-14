<?php

namespace DegonetOvpn\Auth\Services;

use DegonetOvpn\Auth\Utils\DatabaseUtil;
use PDO;

class AuthService
{
    private PDO $conn;

    public function __construct()
    {
        $this->conn = DatabaseUtil::getConn();
    }

    public function login(string $username, string $password): object|null
    {
        $stmt = $this->conn->prepare('SELECT * FROM mikrotik_users WHERE username = ?');
        $stmt->execute([$username]);

        $user = $stmt->fetch(PDO::FETCH_OBJ);

        if (!$user) return null;
        if (!password_verify($password, $user->password)) return null;

        return $user;
    }
}
