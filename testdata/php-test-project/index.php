<?php

/**
 * Entry point for ddev-xdebug-tui manual testing.
 *
 * This file provides a simple call chain across multiple files
 * so the debugger has a meaningful stack to display.
 */

require_once __DIR__ . '/lib/math.php';
require_once __DIR__ . '/lib/greeter.php';

$name  = $_GET['name'] ?? 'world';
$value = 42;

$result  = add_numbers($value, 8);
$message = greet($name, $result);

echo $message . PHP_EOL;
