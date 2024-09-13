<?php

namespace App;

require __DIR__ . '/../vendor/autoload.php';

$logger = new Logger();
$logger->error('testmessage', ['id' => 1]);