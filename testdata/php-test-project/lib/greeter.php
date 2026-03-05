<?php

/**
 * Greeting helper.
 * Set a breakpoint here to inspect $name and $number in a different stack frame.
 */

function greet(string $name, int $number): string
{
    $greeting = "Hello, {$name}! The answer is {$number}.";
    return $greeting;
}
