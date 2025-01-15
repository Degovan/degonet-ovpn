<?php

namespace DegonetOvpn\Auth\Services;

class CCDService
{
    public static function create($user): bool
    {
        $dir = $_ENV['CCD_DIR'];

        if (file_exists("{$dir}/{$user->username}"))
            return false;

        $content = "ifconfig-push {$user->ip} {$user->netmask}\n";
        file_put_contents("{$dir}/{$user->username}", $content);

        return true;
    }
}
