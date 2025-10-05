#!/usr/bin/env pwsh

# Source the main library file
. "$PSScriptRoot/../Lib/OwnerSearch.ps1"

# Run test case for an existing user
Run-TestCase -testName "Test for existing user" -lastName "Ayman" -shouldExist $true

# Close the browser
Close-Browser

# Exit with code 0 to indicate success
exit 0
