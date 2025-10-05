#!/usr/bin/env pwsh

# Install required Selenium PowerShell module
if (-not (Get-Module -ListAvailable -Name Selenium)) {
    Install-Module -Name Selenium -Scope CurrentUser -Force
}

# Import the Selenium module
Import-Module Selenium

# Start a new instance of the Chrome driver using the self-hosted WebDriver server
function Start-SelfHostedWebDriver {
    param (
        [string]$browser = "chrome",
        [string]$seleniumServerUrl = "http://172.17.0.3:4444/wd/hub"
    )

    # Set up the browser-specific capabilities
    if ($browser -eq "chrome") {
        $options = New-Object OpenQA.Selenium.Chrome.ChromeOptions
        #$options.AddArgument("--headless")

        $driver = New-Object OpenQA.Selenium.Remote.RemoteWebDriver -ArgumentList $seleniumServerUrl, $options
    } elseif ($browser -eq "firefox") {
        $options = New-Object OpenQA.Selenium.Firefox.FirefoxOptions

        $driver = New-Object OpenQA.Selenium.Remote.RemoteWebDriver -ArgumentList $seleniumServerUrl, $options
    } else {
        throw "Unsupported browser: $browser"
    }

    return $driver
}

if (-not $env:SELENIUM_SERVER_URL) {
    $global:driver = Start-SelfHostedWebDriver -browser "chrome"
} else {
    $global:driver = Start-SelfHostedWebDriver -browser "chrome" -seleniumServerUrl $env:SELENIUM_SERVER_URL
}

# Function to search for a user by last name
function Search-User {
    param (
        [string]$lastName
    )

    # Get the base URL from environment variable or use the default value
    $baseUrl = $env:APP_URL
    if (-not $baseUrl) {
        $baseUrl = "http://172.17.0.2:8080"
    }

    # Concatenate the full URL
    $fullUrl = "$baseUrl/owners/find"

    # Navigate to the search page
    $global:driver.Navigate().GoToUrl($fullUrl)


    # Wait for the page to load
    Start-Sleep -Seconds 3

    # Find the search bar for last name and enter the last name to search
    $searchBar = $global:driver.FindElementByName("lastName")
    $searchBar.Clear()
    $searchBar.SendKeys($lastName)

    # Find and click the search button
    $searchButton = $global:driver.FindElementByXPath("//button[@type='submit']")
    $searchButton.Click()

    # Wait for the search results to load
    Start-Sleep -Seconds 3

    # Check if the user was found by looking for the presence of the results table
    try {
        $userElement = $global:driver.FindElementByXPath("//h2[text()='Owner Information']")
        return $true
    } catch {
        return $false
    }
}

# Function to handle the test case
function Run-TestCase {
    param (
        [string]$testName,
        [string]$lastName,
        [bool]$shouldExist
    )

    $startTime = Get-Date
    $status = 0

    try {
        # Search for the user
        $userFound = Search-User -lastName $lastName

        if ($userFound) {
            # If the user should exist, this is the expected outcome
            if ($shouldExist) {
                # Extract and print user details (e.g., owner's name)
                try {
                    $ownerName = $global:driver.FindElementByXPath("//td/b")
                    $status = 1
                } catch {
                    $status = 0
                }
            } else {
                $status = 0
            }
        } else {
            # If the user should not exist, this is the expected outcome
            if ($shouldExist) {
                $status = 0
            } else {
                $status = 1
            }
        }
    } catch {
        $status = 0
    }

    # Ensure the browser is closed and the result is reported
    if ($global:driver -ne $null) {
        $global:driver.Quit() | Out-Null
    }
    $duration = (Get-Date) - $startTime
    $data = @{
        test_name = $testName
        status = $status
        duration_seconds = [math]::Round($duration.TotalSeconds, 2)
    }
    $data | ConvertTo-Json | Write-Output
    exit 0
}


# Function to close the browser
function Close-Browser {
    $global:driver.Quit()
}
