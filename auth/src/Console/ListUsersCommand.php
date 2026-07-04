<?php

namespace DegonetOvpn\Auth\Console;

use DegonetOvpn\Auth\Services\AuthService;
use function Laravel\Prompts\intro;
use function Laravel\Prompts\outro;
use function Laravel\Prompts\table;
use function Laravel\Prompts\warning;

class ListUsersCommand
{
    private AuthService $auth;

    public function __construct(AuthService $auth)
    {
        $this->auth = $auth;
    }

    public function run(): int
    {
        $users = $this->auth->listUsers();

        intro('List OpenVPN Users');

        if (count($users) === 0) {
            warning('Belum ada user terdaftar.');
            return 0;
        }

        $rows = [];
        foreach ($users as $user) {
            $rows[] = [
                (string)$user->username,
                (string)$user->ip,
                (string)$user->netmask,
            ];
        }

        table(['Username', 'IP', 'Netmask'], $rows);
        outro('Selesai menampilkan daftar user.');

        return 0;
    }
}
