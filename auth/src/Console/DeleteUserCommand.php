<?php

namespace DegonetOvpn\Auth\Console;

use DegonetOvpn\Auth\Services\AuthService;
use DegonetOvpn\Auth\Services\CCDService;
use DegonetOvpn\Auth\Validator\DeleteUserInputValidator;
use function Laravel\Prompts\error;
use function Laravel\Prompts\intro;
use function Laravel\Prompts\outro;
use function Laravel\Prompts\table;

class DeleteUserCommand
{
    private AuthService $auth;
    private DeleteUserInputValidator $validator;

    public function __construct(AuthService $auth, DeleteUserInputValidator $validator)
    {
        $this->auth = $auth;
        $this->validator = $validator;
    }

    public function run(array $argv): int
    {
        $username = $argv[2] ?? null;
        $errors = $this->validator->validate($username);

        if (in_array('missing_required_arguments', $errors, true)) {
            intro('Delete OpenVPN User');
            error('Usage: ./auth delete <username>');
            echo 'Contoh: ./auth delete budi' . PHP_EOL;
            return 1;
        }

        $user = $this->auth->findUserByUsername((string)$username);

        if (!$user) {
            error("User '{$username}' tidak ditemukan.");
            return 1;
        }

        $deleted = $this->auth->deleteUser((string)$username);

        if (!$deleted) {
            error("Gagal menghapus user '{$username}' dari database.");
            return 1;
        }

        $deletedCcd = CCDService::delete((string)$username);
        $ccdStatus = $deletedCcd ? 'dihapus' : 'tidak ditemukan';

        intro('Delete OpenVPN User');
        table([], [
            ['Username', (string)$username],
            ['Status DB', 'dihapus'],
            ['Status CCD', $ccdStatus],
        ]);
        outro('User berhasil dihapus.');

        return 0;
    }
}
