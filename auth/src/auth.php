<?php

use DegonetOvpn\Auth\Services\AuthService;
use DegonetOvpn\Auth\Services\CCDService;
use DegonetOvpn\Auth\Utils\DatabaseUtil;
use Dotenv\Dotenv;

require __DIR__ . '/../vendor/autoload.php';

$dotenv = Dotenv::createMutable(__DIR__ . '/../');
$dotenv->load();

DatabaseUtil::initConnection();

$auth = new AuthService;
$username = $_SERVER['username'] ?? 'testing';
$password = $_SERVER['password'] ?? 'testing123';

if ($user = $auth->login($username, $password)) {
    CCDService::create($user);

    echo 'Login successful' . PHP_EOL;
    exit(0);
} else {
    echo 'Login failed' . PHP_EOL;
    exit(1);
}
