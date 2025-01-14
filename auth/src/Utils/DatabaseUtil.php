<?php

namespace DegonetOvpn\Auth\Utils;

use Exception;
use PDO;

class DatabaseUtil
{
    private static PDO $connection;

    public static function initConnection()
    {
        try {
            $pdo = new PDO(
                'mysql:host=' . $_ENV['DB_HOST'] . ';dbname=' . $_ENV['DB_NAME'],
                $_ENV['DB_USER'],
                $_ENV['DB_PASS']
            );

            $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
            static::$connection = $pdo;
            return $pdo;
        } catch (Exception $e) {
            echo 'Database connection error: ' . $e->getMessage();
            exit(1);
        }
    }

    public static function getConn()
    {
        if (static::$connection) return static::$connection;
        return static::initConnection();
    }
}
