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

    public function addUser(string $username, string $password, string $ip, string $netmask): object|null
    {
        $checkStmt = $this->conn->prepare('SELECT id FROM mikrotik_users WHERE username = ?');
        $checkStmt->execute([$username]);

        if ($checkStmt->fetch()) {
            return null;
        }

        $checkIpStmt = $this->conn->prepare('SELECT id FROM mikrotik_users WHERE ip = ?');
        $checkIpStmt->execute([$ip]);

        if ($checkIpStmt->fetch()) {
            return null;
        }

        $insertStmt = $this->conn->prepare(
            'INSERT INTO mikrotik_users (username, password, ip, netmask) VALUES (?, ?, ?, ?)'
        );

        $insertStmt->execute([
            $username,
            password_hash($password, PASSWORD_DEFAULT),
            $ip,
            $netmask,
        ]);

        $stmt = $this->conn->prepare('SELECT * FROM mikrotik_users WHERE id = ?');
        $stmt->execute([(int)$this->conn->lastInsertId()]);

        $user = $stmt->fetch(PDO::FETCH_OBJ);
        return $user ?: null;
    }

    public function findUserByUsername(string $username): object|null
    {
        $stmt = $this->conn->prepare('SELECT * FROM mikrotik_users WHERE username = ?');
        $stmt->execute([$username]);

        $user = $stmt->fetch(PDO::FETCH_OBJ);
        return $user ?: null;
    }

    public function findUserByIp(string $ip): object|null
    {
        $stmt = $this->conn->prepare('SELECT * FROM mikrotik_users WHERE ip = ?');
        $stmt->execute([$ip]);

        $user = $stmt->fetch(PDO::FETCH_OBJ);
        return $user ?: null;
    }

    /**
     * @return array<int, object>
     */
    public function listUsers(): array
    {
        $stmt = $this->conn->query('SELECT username, ip, netmask FROM mikrotik_users ORDER BY username ASC');
        return $stmt->fetchAll(PDO::FETCH_OBJ) ?: [];
    }

    public function deleteUser(string $username): bool
    {
        $stmt = $this->conn->prepare('DELETE FROM mikrotik_users WHERE username = ?');
        $stmt->execute([$username]);

        return $stmt->rowCount() > 0;
    }
}
