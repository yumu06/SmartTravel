param(
    [ValidateSet('all', 'authorization', 'user_info', 'post_detail', 'recommendation')]
    [string]$Scenario = 'all'
)

$ErrorActionPreference = 'Stop'

foreach ($required in @('BENCHMARK_MYSQL_USER', 'BENCHMARK_MYSQL_PASSWORD', 'BENCHMARK_REDIS_PASSWORD')) {
    if ([string]::IsNullOrWhiteSpace([Environment]::GetEnvironmentVariable($required))) {
        throw "$required is required"
    }
}

$benchAccess = [guid]::NewGuid().ToString('N') + [guid]::NewGuid().ToString('N')
$benchRefresh = [guid]::NewGuid().ToString('N') + [guid]::NewGuid().ToString('N')
$env:BENCHMARK_JWT_ACCESS_SECRET = $benchAccess
$env:BENCHMARK_JWT_REFRESH_SECRET = $benchRefresh
$env:BENCHMARK_REDIS_ADDR = '127.0.0.1:6379'

go build -o benchmark\travel-benchmark.exe .
go build -o benchmark\benchmark-token.exe .\cmd\benchmark-token
go build -o benchmark\loadtest.exe .\cmd\loadtest

$env:BENCHMARK_ACCESS_TOKEN = (& .\benchmark\benchmark-token.exe).Trim()
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($env:BENCHMARK_ACCESS_TOKEN)) {
    throw 'benchmark token creation failed'
}

$env:SERVER_PORT = '1017'
$env:SERVER_ACCESS_LOG = 'false'
$env:GIN_MODE = 'release'
$env:MYSQL_DATABASE = 'travel_benchmark'
$env:MYSQL_ROOT = $env:BENCHMARK_MYSQL_USER
$env:MYSQL_PASSWORD = $env:BENCHMARK_MYSQL_PASSWORD
$env:JWT_ACCESS_SECRET = $benchAccess
$env:JWT_REFRESH_SECRET = $benchRefresh
$env:JWT_ISSUER = 'ongoing-trip-benchmark'
$env:REDIS_ADDR = '127.0.0.1:6379'
$env:REDIS_PASSWORD = $env:BENCHMARK_REDIS_PASSWORD

$benchmarkDir = (Resolve-Path .\benchmark).Path
$server = Start-Process -FilePath (Resolve-Path .\benchmark\travel-benchmark.exe) `
    -WindowStyle Hidden -PassThru `
    -RedirectStandardOutput (Join-Path $benchmarkDir 'server-1017.out.log') `
    -RedirectStandardError (Join-Path $benchmarkDir 'server-1017.err.log')

$postID = (Get-Content .\benchmark\post_ids.txt -TotalCount 1).Trim()
$headers = @{ Authorization = 'Bearer ' + $env:BENCHMARK_ACCESS_TOKEN }
$smokeUrls = @(
    'http://127.0.0.1:1017/travel/authorization',
    'http://127.0.0.1:1017/travel/user/info',
    ('http://127.0.0.1:1017/travel/post/show/' + $postID),
    'http://127.0.0.1:1017/travel/post/recommand?limit=20'
)

$ready = $false
for ($attempt = 0; $attempt -lt 50; $attempt++) {
    try {
        $response = Invoke-WebRequest -UseBasicParsing -Uri $smokeUrls[0] -Headers $headers -TimeoutSec 2
        if ($response.StatusCode -eq 200) {
            $ready = $true
            break
        }
    } catch {
        Start-Sleep -Milliseconds 200
    }
}
if (-not $ready) {
    throw 'benchmark server did not become ready'
}

foreach ($smokeUrl in $smokeUrls) {
    $response = Invoke-WebRequest -UseBasicParsing -Uri $smokeUrl -Headers $headers -TimeoutSec 5
    Write-Output "SMOKE status=$($response.StatusCode) url=$smokeUrl"
}

$postUrls = Get-Content .\benchmark\post_ids.txt | ForEach-Object {
    'http://127.0.0.1:1017/travel/post/show/' + $_.Trim()
}
[IO.File]::WriteAllLines((Join-Path $benchmarkDir 'post_urls.txt'), $postUrls)

function Run-Scenario([string]$Name, [string]$Url, [string]$UrlsFile = '') {
    foreach ($concurrency in @(10, 50, 100, 200)) {
        $requests = if ($concurrency -eq 10) { 500 } else { 5000 }
        Write-Output "SCENARIO=$Name CONCURRENCY=$concurrency"
        if ($UrlsFile) {
            $lines = & .\benchmark\loadtest.exe -urls-file $UrlsFile -n $requests -c $concurrency -timeout 5s
        } else {
            $lines = & .\benchmark\loadtest.exe -url $Url -n $requests -c $concurrency -timeout 5s
        }
        $lines | Write-Output
        $summary = $lines | Select-Object -First 1
        if ($summary -match 'error_rate=([0-9.]+)' -and [double]$Matches[1] -ge 0.01) {
            Write-Output "STOP=$Name error_rate=$($Matches[1])"
            break
        }
    }
}

if ($Scenario -in @('all', 'authorization')) {
    Run-Scenario 'authorization' 'http://127.0.0.1:1017/travel/authorization'
}
if ($Scenario -in @('all', 'user_info')) {
    Run-Scenario 'user_info' 'http://127.0.0.1:1017/travel/user/info'
}
if ($Scenario -in @('all', 'post_detail')) {
    Run-Scenario 'post_detail' '' '.\benchmark\post_urls.txt'
}
if ($Scenario -in @('all', 'recommendation')) {
    Run-Scenario 'recommendation' 'http://127.0.0.1:1017/travel/post/recommand?limit=20'
}

$health = Invoke-WebRequest -UseBasicParsing -Uri $smokeUrls[0] -Headers $headers -TimeoutSec 5
Write-Output "FINAL_HEALTH status=$($health.StatusCode) benchmark_pid=$($server.Id)"
