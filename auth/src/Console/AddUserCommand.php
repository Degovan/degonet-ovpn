<?php

namespace DegonetOvpn\Auth\Console;

use DegonetOvpn\Auth\Services\AuthService;
use DegonetOvpn\Auth\Services\CCDService;
use DegonetOvpn\Auth\Validator\AddUserInputValidator;
use function Laravel\Prompts\error;
use function Laravel\Prompts\intro;
use function Laravel\Prompts\outro;
use function Laravel\Prompts\table;
use function Laravel\Prompts\warning;

class AddUserCommand
{
    private AuthService $auth;
    private AddUserInputValidator $validator;

    public function __construct(AuthService $auth, AddUserInputValidator $validator)
    {
        $this->auth = $auth;
        $this->validator = $validator;
    }

    public function run(array $argv): int
    {
        $newUsername = $argv[2] ?? null;
        $newIp = $argv[3] ?? null;
        $newPassword = $argv[4] ?? null;
        $defaultNetmask = $_ENV['DEFAULT_NETMASK'] ?? '255.255.255.0';
        $newNetmask = $argv[5] ?? $defaultNetmask;

        $errors = $this->validator->validate($newUsername, $newIp, $newNetmask);

        if (in_array('missing_required_arguments', $errors, true)) {
            intro('Add OpenVPN User');
            error('Usage: ./auth add <username> <ip> [password] [netmask]');
            echo 'Contoh: ./auth add budi 10.8.0.10 rahasia 255.255.255.0' . PHP_EOL;
            return 1;
        }

        foreach ($errors as $error) {
            error($error);
            return 1;
        }

        if (!$newPassword) {
            $newPassword = $newUsername;
        }

        $existingByUsername = $this->auth->findUserByUsername($newUsername);
        if ($existingByUsername) {
            error("User '{$newUsername}' sudah ada.");
            return 1;
        }

        $existingByIp = $this->auth->findUserByIp($newIp);
        if ($existingByIp) {
            error("IP '{$newIp}' sudah dipakai oleh user '{$existingByIp->username}'.");
            return 1;
        }

        $user = $this->auth->addUser($newUsername, $newPassword, $newIp, $newNetmask);

        if (!$user) {
            error('Gagal menambahkan user. Pastikan username dan IP belum terpakai.');
            return 1;
        }

        $createdCcd = CCDService::create($user);
        $ccdStatus = $createdCcd ? 'dibuat' : 'sudah ada (skip)';

        intro('Add OpenVPN User');
        table(
            [],
            [
                ['Username', (string)$user->username],
                ['IP', (string)$user->ip],
                ['Netmask', (string)$user->netmask],
                ['Password', '****'],
                ['Status CCD', $ccdStatus],
            ]
        );
        outro('User berhasil ditambahkan.');

        if (!$createdCcd) {
            warning('User DB berhasil ditambah, tapi file CCD sudah ada sehingga tidak dibuat ulang.');
        }

        return 0;
    }
}
