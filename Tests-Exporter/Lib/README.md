Sure, here is a detailed documentation for the PowerShell script. This documentation includes the necessary environment variables, detailed instructions on how to use and run the script, and what each part of the script does.

# PowerShell Selenium Test Script Documentation

## Overview

This PowerShell script is designed to automate the process of testing a web application using Selenium WebDriver. It supports running tests on both Chrome and Firefox browsers, and it outputs the results in JSON format.

## Environment Variables

The script relies on certain environment variables to configure its behavior. These environment variables must be set before running the script:

- `SELENIUM_SERVER_URL`: The URL of the Selenium server. If not set, the script defaults to `http://localhost:4444/wd/hub`.
- `APP_URL`: The base URL of the application under test. If not set, the script defaults to `http://172.17.0.3:8080`.

## Script Structure

### 1. Install Required Selenium PowerShell Module

The script begins by checking if the Selenium module is installed. If not, it installs the module:

```powershell
if (-not (Get-Module -ListAvailable -Name Selenium)) {
    Install-Module -Name Selenium -Scope CurrentUser -Force | Out-Null
}
```

### 2. Import the Selenium Module

The Selenium module is then imported:

```powershell
Import-Module Selenium | Out-Null
```

### 3. Start the WebDriver

The `Start-SelfHostedWebDriver` function initializes a WebDriver instance for the specified browser (Chrome or Firefox). It fetches the Selenium server URL from the environment variable `SELENIUM_SERVER_URL` or defaults to `http://localhost:4444/wd/hub` if not set:

```powershell
function Start-SelfHostedWebDriver {
    param (
        [string]$browser = "chrome"
    )

    $seleniumServerUrl = $env:SELENIUM_SERVER_URL
    if (-not $seleniumServerUrl) {
        $seleniumServerUrl = "http://localhost:4444/wd/hub"
    }

    if ($browser -eq "chrome") {
        $options = New-Object OpenQA.Selenium.Chrome.ChromeOptions
        $driver = New-Object OpenQA.Selenium.Remote.RemoteWebDriver -ArgumentList $seleniumServerUrl, $options
    } elseif ($browser -eq "firefox") {
        $options = New-Object OpenQA.Selenium.Firefox.FirefoxOptions
        $driver = New-Object OpenQA.Selenium.Remote.RemoteWebDriver -ArgumentList $seleniumServerUrl, $options
    } else {
        $null = $null
        return $null
    }

    return $driver
}

$global:driver = Start-SelfHostedWebDriver -browser "chrome"
```

### 4. Search for a User

The `Search-User` function searches for a user by their last name on the application. It navigates to the search page, enters the last name, and submits the search form:

```powershell
function Search-User {
    param (
        [string]$lastName
    )

    $baseUrl = $env:APP_URL
    if (-not $baseUrl) {
        $baseUrl = "http://172.17.0.3:8080"
    }

    $fullUrl = "$baseUrl/owners/find"
    $global:driver.Navigate().GoToUrl($fullUrl)
    Start-Sleep -Seconds 3

    $searchBar = $global:driver.FindElementByName("lastName")
    $searchBar.Clear()
    $searchBar.SendKeys($lastName)

    $searchButton = $global:driver.FindElementByXPath("//button[@type='submit']")
    $searchButton.Click()
    Start-Sleep -Seconds 3

    try {
        $userElement = $global:driver.FindElementByXPath("//h2[text()='Owner Information']")
        return $true
    } catch {
        return $false
    }
}
```

### 5. Run Test Cases

The `Run-TestCase` function runs a test case to search for a user and checks if the user exists or not. The result is outputted in JSON format:

```powershell
function Run-TestCase {
    param (
        [string]$testName,
        [string]$lastName,
        [bool]$shouldExist
    )

    $startTime = Get-Date
    $status = "failure"

    try {
        $userFound = Search-User -lastName $lastName

        if ($userFound) {
            if ($shouldExist) {
                try {
                    $ownerName = $global:driver.FindElementByXPath("//td/b")
                    $status = "success"
                } catch {
                    $status = "failure"
                }
            } else {
                $status = "failure"
            }
        } else {
            if ($shouldExist) {
                $status = "failure"
            } else {
                $status = "success"
            }
        }
    } catch {
        $status = "failure"
    }

    if ($global:driver -ne $null) {
        $global:driver.Quit() | Out-Null
    }
    $duration = (Get-Date) - $startTime
    $data = @{
        test_name = $testName
        status = $status
        duration_seconds = [math]::Round($duration.TotalSeconds, 2)
    }
    $data | ConvertTo-Json
    exit 0
}
```

### 6. Close the Browser

The `Close-Browser` function gracefully closes the browser:

```powershell
function Close-Browser {
    if ($global:driver -ne $null) {
        $global:driver.Quit() | Out-Null
    }
}
```

### 7. Example Test Case Execution

Example test cases are provided to demonstrate how to run the tests:

```powershell
Run-TestCase -testName "Test for existing user Ayman" -lastName "Ayman" -shouldExist $true
Run-TestCase -testName "Test for non-existing user Smith" -lastName "Smith" -shouldExist $false

Close-Browser
```

## How to Use and Run the Script

1. **Set Environment Variables**:
   - `SELENIUM_SERVER_URL`: The URL of the Selenium server (e.g., `http://localhost:4444/wd/hub`).
   - `APP_URL`: The base URL of the application under test (e.g., `http://172.17.0.3:8080`).

2. **Run the Script**:
   - Open a PowerShell terminal.
   - Set the necessary environment variables:

     ```powershell
     $env:SELENIUM_SERVER_URL = "http://your-selenium-server-url:4444/wd/hub"
     $env:APP_URL = "http://your-application-url:8080"
     ```

   - Execute the script:

     ```powershell
     ./your-script-name.ps1
     ```

3. **Interpreting the Output**:
   - The script will output the result of each test case in JSON format. For example:

     ```json
     {
         "test_name": "Test for existing user Ayman",
         "status": "success",
         "duration_seconds": 5.34
     }
     ```

This documentation provides a detailed explanation of how to set up, use, and run the PowerShell script for Selenium WebDriver testing. Make sure to replace `your-selenium-server-url` and `your-application-url` with the actual URLs for your setup.
