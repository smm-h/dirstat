package main

// Exit codes returned by command handlers.
const (
	ExitSuccess = 0 // command completed successfully
	ExitGeneral = 1 // runtime failure (scan error, output error)
	ExitUsage   = 2 // invalid input: bad flag value or bad target path
)
