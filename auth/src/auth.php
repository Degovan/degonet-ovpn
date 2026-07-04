<?php

use DegonetOvpn\Auth\Console\AddUserCommand;
use DegonetOvpn\Auth\Console\DeleteUserCommand;
use DegonetOvpn\Auth\Console\ListUsersCommand;
use DegonetOvpn\Auth\Services\AuthService;
use DegonetOvpn\Auth\Services\CCDService;
use DegonetOvpn\Auth\Utils\DatabaseUtil;
use DegonetOvpn\Auth\Validator\AddUserInputValidator;
use DegonetOvpn\Auth\Validator\DeleteUserInputValidator;
use Dotenv\Dotenv;
use function Laravel\Prompts\error;
use function Laravel\Prompts\intro;

require __DIR__ . '/../vendor/autoload.php';

$dotenv = Dotenv::createMutable(__DIR__ . '/../');
$dotenv->load();

DatabaseUtil::initConnection();

$auth = new AuthService;

if (!isset($_SERVER['username']) && PHP_SAPI === 'cli') {
    $commandName = $argv[1] ?? null;

    if ($commandName === 'add') {
        $command = new AddUserCommand($auth, new AddUserInputValidator);
        exit($command->run($argv));
    }

    if ($commandName === 'delete') {
        $command = new DeleteUserCommand($auth, new DeleteUserInputValidator);
        exit($command->run($argv));
    }

    if ($commandName === 'list') {
        $command = new ListUsersCommand($auth);
        exit($command->run());
    }

    intro('OpenVPN Auth CLI');
    error('Command tidak dikenal. Gunakan salah satu: add, delete, list');
    echo 'Contoh:' . PHP_EOL;
    echo '  ./auth add <username> <ip> [password] [netmask]' . PHP_EOL;
    echo '  ./auth delete <username>' . PHP_EOL;
    echo '  ./auth list' . PHP_EOL;
    exit(1);
}

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
