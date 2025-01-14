<?php

if (file_exists('auth.phar'))
    unlink('auth.phar');

$phar = new Phar('auth.phar');
$phar->buildFromDirectory(__DIR__);
$phar->setStub(file_get_contents('phar.php'));

echo "Phar compiled successfully\n";

rename('auth.phar', '../auth.phar');
echo "Phar moved to parent directory\n";
