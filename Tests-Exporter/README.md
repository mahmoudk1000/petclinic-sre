# PetClinic Test Monitoring App

This application is designed to monitor the health of a web application (such as a PetClinic app) by running automated tests using PowerShell scripts and exposing the test results as Prometheus metrics. It utilizes Selenium to perform browser-based tests and reports success or failure, along with the duration of each test.

## Features

- Executes PowerShell scripts located in the `Tests` directory.
- Reports the status (success or failure) and duration of each test as Prometheus metrics.
- Runs tests on a set interval (default: every 5 minutes).
- Exposes the test results to Prometheus via an HTTP endpoint.

## Components

- **Go Application**: Handles running tests and exposing Prometheus metrics.
- **PowerShell Scripts**: Perform Selenium-based tests on the PetClinic app.
- **Prometheus Metrics**: Exposes test results and durations for monitoring.

## Prerequisites

Before running the application, make sure the following prerequisites are met:

### 1. Go Language

Ensure that Go is installed on your system. You can download and install Go from [here](https://go.dev/dl/).

### 2. PowerShell

The app requires PowerShell to run the test scripts. PowerShell should be available and configured for the system. The script also requires the `Selenium` module to interact with the browser.

### 3. Selenium WebDriver

This app uses Selenium for browser automation. Ensure that the Selenium WebDriver is available and properly configured. The PowerShell scripts expect the WebDriver to be available at a specified URL (e.g., `http://172.17.0.3:4444/wd/hub`).

### 4. Prometheus

Ensure that Prometheus is installed and configured to scrape metrics from the Go application.

## Environment Variables

The application relies on a few environment variables to configure its behavior:

### Required Variables

- **`APP_URL`**: The base URL of the application under test (default: `http://localhost:8080`).
  
Example:

```sh
  export APP_URL="http://your-app-url.com"

    SELENIUM_SERVER_URL: The URL where the Selenium WebDriver is hosted (default: http://172.17.0.3:4444/wd/hub).

    Example:

    export SELENIUM_SERVER_URL="http://your-selenium-server-url:4444/wd/hub"
```

Optional Variables:

    PORT: The port on which the Go application will expose metrics (default: 9091).

    Example:

    export PORT="9091"

Directory Structure

```plain
/your-app-directory
├── main.go          # Main Go application that runs the tests and exposes metrics
├── Tests            # Directory containing PowerShell test scripts
│   ├── test1.ps1    # PowerShell test script 1
│   └── test2.ps1    # PowerShell test script 2
├── Lib              # Directory containing the PowerShell helper libraries
│   └── OwnerSearch.ps1   # Selenium PowerShell script for user search
└── README.md        # This file
```

Test Directory

Place your PowerShell test scripts inside the Tests directory. Each test script should:

    Execute a test case (using functions defined in the Lib directory).
    Report the result in JSON format (e.g., test name, status, duration).

Example test (test1.ps1):

```PowerShell
#!/usr/bin/env pwsh

. "$PSScriptRoot/../Lib/OwnerSearch.ps1"

Run-TestCase -testName "Test for non-existing user" -lastName "CityStars" -shouldExist $false

Close-Browser

exit 0
```

Running the Application
Step 1: Clone the repository

Clone this repository to your local machine:

git clone <https://your-repository-url.git>
cd your-repository-directory

Step 2: Set up the environment variables

Set up the required environment variables. Example:

export APP_URL="<http://localhost:8080>"
export SELENIUM_SERVER_URL="<http://172.17.0.3:4444/wd/hub>"
export PORT="9091"

Step 3: Install Go dependencies

Ensure that the required Go dependencies are installed:

go mod tidy

Step 4: Build and Run the Application

Build and run the Go application:

go run main.go

This will start the Go application, which will:

    Begin running tests from the Tests directory immediately.
    Run tests on a 5-minute interval.
    Expose Prometheus metrics at http://localhost:9091/metrics.

Step 5: Set up Prometheus

Configure Prometheus to scrape the metrics exposed by the Go application. Example configuration:

scrape_configs:

- job_name: 'petclinic-tests'
    static_configs:
  - targets: ['localhost:9091']

Step 6: View the Metrics

You can now view the test metrics in Prometheus. The metrics include:

    petclinic_test_status: The status of the test (1 for success, 0 for failure).
    petclinic_test_duration_seconds: The duration of each test in seconds.

Metrics Example

# HELP petclinic_test_status Status of the test (1 = success, 0 = failure)

# TYPE petclinic_test_status gauge

petclinic_test_status{test_name="Test for non-existing user"} 0

# HELP petclinic_test_duration_seconds Duration of the test in seconds

# TYPE petclinic_test_duration_seconds gauge

petclinic_test_duration_seconds{test_name="Test for non-existing user"} 12.5

Test Execution Flow

    The Go application starts and begins executing all PowerShell scripts found in the Tests directory.
    Each test script runs a test case using the functions defined in the Lib directory.
    The test results (status and duration) are collected and sent to Prometheus via HTTP.
    Tests are rerun every 5 minutes to ensure the application is continuously monitored.

Troubleshooting

    Error running PowerShell script: Ensure that PowerShell is installed and configured properly on your system.
    Selenium errors: Ensure that the Selenium WebDriver is running and accessible at the configured URL (SELENIUM_SERVER_URL).
    Prometheus issues: Verify that Prometheus is correctly scraping the /metrics endpoint of the Go application.
