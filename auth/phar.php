#!/usr/bin/php
<?php

Phar::mapPhar('auth.phar');
require 'phar://auth.phar/src/auth.php';
__HALT_COMPILER();
