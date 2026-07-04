<?php

namespace DegonetOvpn\Auth\Validator;

class AddUserInputValidator
{
    /**
     * @return string[]
     */
    public function validate(?string $username, ?string $ip, ?string $netmask): array
    {
        $errors = [];

        if (!$username || !$ip) {
            $errors[] = 'missing_required_arguments';
            return $errors;
        }

        if (!$this->isValidIpv4($ip)) {
            $errors[] = 'IP tidak valid. Gunakan format IPv4, contoh: 10.8.0.10';
        }

        if (!$this->isValidIpv4((string)$netmask)) {
            $errors[] = 'Netmask tidak valid. Gunakan format IPv4, contoh: 255.255.255.0';
        }

        return $errors;
    }

    private function isValidIpv4(string $value): bool
    {
        return (bool)filter_var($value, FILTER_VALIDATE_IP, FILTER_FLAG_IPV4);
    }
}
