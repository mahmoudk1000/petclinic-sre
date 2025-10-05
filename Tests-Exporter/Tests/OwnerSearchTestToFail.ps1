#!/usr/bin/env pwsh

# Source the main library file
. "$PSScriptRoot/../Lib/OwnerSearch.ps1"

# Run test case for a non-existing user
Run-TestCase -testName "Test for non-existing user" -lastName "CityStars" -shouldExist $false

# Close the browser
Close-Browser

# Exit with code 0 to indicate success
exit 0
