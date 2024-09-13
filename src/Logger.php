<?php

namespace App;

require __DIR__ . '/../vendor/autoload.php';

use Monolog\Formatter\JsonFormatter;
use Monolog\Handler\RotatingFileHandler;
use Monolog\Handler\SocketHandler;
use Monolog\Level;
use Monolog\Logger as MonologLogger;

class Logger implements LoggerInterface
{
    private function send($level, $message, $context = []): void
    {
        $port = getenv()['GOLOGGER_PORT'];
        $logger = new MonologLogger('my_logger');

        $fileHandler = new RotatingFileHandler(__DIR__ . "/log/debug.log", 5);
        $socketHandler = new SocketHandler("udp://localhost:$port");

        $jsonFormatter = new JsonFormatter();
        $fileHandler->setFormatter($jsonFormatter);
        $socketHandler->setFormatter($jsonFormatter);

        $logger->pushHandler($fileHandler);
        $logger->pushHandler($socketHandler);

        $name = $level->name;
        $logger->$name($message, $context);
    }

    public function emergency($message, array $context = []): void
    {
        $this->send(Level::Emergency, $message, $context);
    }

    public function alert($message, array $context = []): void
    {
        $this->send(Level::Alert, $message, $context);
    }

    public function critical($message, array $context = []): void
    {
        $this->send(Level::Critical, $message, $context);
    }

    public function error($message, array $context = []): void
    {
        $this->send(Level::Error, $message, $context);
    }

    public function warning($message, array $context = []): void
    {
        $this->send(Level::Warning, $message, $context);
    }

    public function notice($message, array $context = []): void
    {
        $this->send(Level::Notice, $message, $context);
    }

    public function info($message, array $context = []): void
    {
        $this->send(Level::Info, $message, $context);
    }

    public function debug($message, array $context = []): void
    {
        $this->send(Level::Debug, $message, $context);
    }
}
