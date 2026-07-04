<?php

namespace DegonetOvpn\Auth\Utils;

use Exception;
use PDO;
use Phar;

class DatabaseUtil
{
    private static ?PDO $connection = null;

    public static function initConnection()
    {
        try {
            $dbFile = static::resolveDbFilePath($_ENV['DB_FILE'] ?? 'users.sqlite');
            static::ensureDbFileExists($dbFile);

            $pdo = new PDO('sqlite:' . $dbFile);

            $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

            $pdo->exec('CREATE TABLE IF NOT EXISTS mikrotik_users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT NOT NULL UNIQUE,
                password TEXT NOT NULL,
                ip TEXT NOT NULL,
                netmask TEXT NOT NULL
            )');

            static::$connection = $pdo;
            return $pdo;
        } catch (Exception $e) {
            echo 'Database connection error: ' . $e->getMessage();
            exit(1);
        }
    }

    public static function getConn()
    {
        if (static::$connection instanceof PDO) return static::$connection;
        return static::initConnection();
    }

    private static function resolveDbFilePath(string $dbFile): string
    {
        $dbFile = trim($dbFile);
        if ($dbFile === '') {
            $dbFile = 'users.sqlite';
        }

        if (static::isAbsolutePath($dbFile)) {
            return $dbFile;
        }

        $runningPhar = Phar::running(false);
        if ($runningPhar !== '') {
            return dirname($runningPhar) . DIRECTORY_SEPARATOR . ltrim($dbFile, '/\\');
        }

        return dirname(__DIR__, 2) . DIRECTORY_SEPARATOR . ltrim($dbFile, '/\\');
    }

    private static function ensureDbFileExists(string $dbFile): void
    {
        $dbDir = dirname($dbFile);
        if (!is_dir($dbDir) && !mkdir($dbDir, 0777, true) && !is_dir($dbDir)) {
            throw new Exception('Cannot create database directory: ' . $dbDir);
        }

        if (!file_exists($dbFile) && false === touch($dbFile)) {
            throw new Exception('Cannot create database file: ' . $dbFile);
        }
    }

    private static function isAbsolutePath(string $path): bool
    {
        if ($path === '') {
            return false;
        }

        return $path[0] === '/' || preg_match('/^[A-Za-z]:[\\\\\/]/', $path) === 1;
    }
}
