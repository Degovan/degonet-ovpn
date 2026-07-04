<?php

namespace DegonetOvpn\Auth\Validator;

class DeleteUserInputValidator
{
    /**
     * @return string[]
     */
    public function validate(?string $username): array
    {
        if (!$username) {
            return ['missing_required_arguments'];
        }

        return [];
    }
}
