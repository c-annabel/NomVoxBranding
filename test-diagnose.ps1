# test-diagnose.ps1
# Kills any running server, rebuilds, starts fresh, hits /api/diagnose, pretty-prints result.
# Run from workspace root:  .\test-diagnose.ps1

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$ServerExe  = ".\nomvox-server.exe"
$DiagnoseURL = "http://localhost:8080/api/diagnose"
$PingURL     = "http://localhost:8080/api/ping"

# ── 1. Kill any existing server ───────────────────────────────────
Write-Host ""
Write-Host "──────────────────────────────────────────" -ForegroundColor DarkGray
Write-Host " STEP 1 : Kill old server" -ForegroundColor White
Write-Host "──────────────────────────────────────────" -ForegroundColor DarkGray
$old = Get-Process -Name "nomvox-server" -ErrorAction SilentlyContinue
if ($old) {
    $old | Stop-Process -Force
    Write-Host "  Killed existing nomvox-server (PID $($old.Id))" -ForegroundColor Yellow
    Start-Sleep -Milliseconds 600
} else {
    Write-Host "  No existing server process found — OK" -ForegroundColor DarkGray
}

# ── 2. Rebuild ────────────────────────────────────────────────────
Write-Host ""
Write-Host "──────────────────────────────────────────" -ForegroundColor DarkGray
Write-Host " STEP 2 : go build" -ForegroundColor White
Write-Host "──────────────────────────────────────────" -ForegroundColor DarkGray
$build = & go build -o $ServerExe ./cmd/server/ 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "  BUILD FAILED:" -ForegroundColor Red
    Write-Host $build -ForegroundColor Red
    exit 1
}
Write-Host "  Build OK" -ForegroundColor Green

# ── 3. Start server ───────────────────────────────────────────────
Write-Host ""
Write-Host "──────────────────────────────────────────" -ForegroundColor DarkGray
Write-Host " STEP 3 : Start server" -ForegroundColor White
Write-Host "──────────────────────────────────────────" -ForegroundColor DarkGray
Start-Process -FilePath $ServerExe -WindowStyle Hidden
Write-Host "  Server starting…"
Start-Sleep -Seconds 3

# Wait for /api/ping (up to 10 seconds)
$ready = $false
for ($i = 0; $i -lt 10; $i++) {
    try {
        $ping = Invoke-WebRequest -Uri $PingURL -UseBasicParsing -TimeoutSec 2
        if ($ping.StatusCode -eq 200) { $ready = $true; break }
    } catch {}
    Start-Sleep -Seconds 1
}
if (-not $ready) {
    Write-Host "  Server did not respond to /api/ping after 10s — check for port conflicts" -ForegroundColor Red
    exit 1
}
Write-Host "  Server is up (/api/ping responded 200)" -ForegroundColor Green

# ── 4. Hit /api/diagnose ──────────────────────────────────────────
Write-Host ""
Write-Host "──────────────────────────────────────────" -ForegroundColor DarkGray
Write-Host " STEP 4 : GET $DiagnoseURL" -ForegroundColor White
Write-Host "──────────────────────────────────────────" -ForegroundColor DarkGray

try {
    $resp = Invoke-WebRequest -Uri $DiagnoseURL -UseBasicParsing -TimeoutSec 60
    $json = $resp.Content | ConvertFrom-Json
} catch {
    Write-Host "  HTTP ERROR: $_" -ForegroundColor Red
    exit 1
}

# ── 5. Pretty-print results ───────────────────────────────────────
Write-Host ""
Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  DIAGNOSE RESULTS" -ForegroundColor Cyan
Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan

# ── ENV ──
Write-Host ""
Write-Host "  [ENV]" -ForegroundColor White
$env = $json.env
Write-Host ("    WATSONX_API_KEY    : " + $env.WATSONX_API_KEY)
Write-Host ("    WATSONX_PROJECT_ID : " + $env.WATSONX_PROJECT_ID)
Write-Host ("    WATSONX_URL        : " + $env.WATSONX_URL)
Write-Host ("    WATSONX_CPD_URL    : " + $env.WATSONX_CPD_URL)
Write-Host ("    REDIS_URL          : " + $env.REDIS_URL)
Write-Host ("    CPD_mode           : " + $env.CPD_mode)

# ── TOKEN ──
Write-Host ""
Write-Host "  [TOKEN EXCHANGE]" -ForegroundColor White
$tok = $json.token_exchange
$tokColor = if ($tok.status -eq "OK") { "Green" } else { "Red" }
Write-Host ("    Status  : " + $tok.status) -ForegroundColor $tokColor
if ($tok.status -eq "OK") {
    Write-Host ("    Length  : " + $tok.token_length)
    Write-Host ("    Prefix  : " + $tok.token_prefix)
} elseif ($tok.error) {
    Write-Host ("    Error   : " + $tok.error) -ForegroundColor Red
}

# ── GRANITE CALL ──
Write-Host ""
Write-Host "  [GRANITE CALL]" -ForegroundColor White
$gra = $json.granite_call
$graColor = if ($gra.status -eq "OK") { "Green" } else { "Red" }
Write-Host ("    Status  : " + $gra.status) -ForegroundColor $graColor
if ($gra.status -eq "OK") {
    Write-Host ("    Response length : " + $gra.response_length) -ForegroundColor Green
    Write-Host ("    Response prefix : " + $gra.response_prefix) -ForegroundColor Cyan
} elseif ($gra.error) {
    Write-Host ("    Error   : " + $gra.error) -ForegroundColor Red
}

# ── SUMMARY ──
Write-Host ""
Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan
$tokenOK  = $tok.status -eq "OK"
$graniteOK = $gra.status -eq "OK"

if ($tokenOK -and $graniteOK) {
    Write-Host "  ✓  ALL CLEAR — Granite is working end-to-end!" -ForegroundColor Green
    Write-Host "  Run  .\test-generate.ps1  to test full name generation." -ForegroundColor Green
} elseif ($tokenOK -and -not $graniteOK) {
    Write-Host "  ✓  Token OK   ✗  Granite call failed" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  Most likely fix:" -ForegroundColor White
    Write-Host "  1. Go to https://dataplatform.cloud.ibm.com/projects" -ForegroundColor Yellow
    Write-Host "  2. Open your NomVox project → Manage → Services & integrations" -ForegroundColor Yellow
    Write-Host "  3. Click 'Associate service' → Watson Machine Learning → Confirm" -ForegroundColor Yellow
    Write-Host "  4. If no WML instance exists:" -ForegroundColor Yellow
    Write-Host "     https://cloud.ibm.com/catalog/services/watson-machine-learning" -ForegroundColor Yellow
    Write-Host "     → Lite (free) plan → Create → then associate" -ForegroundColor Yellow
} elseif (-not $tokenOK) {
    Write-Host "  ✗  Token exchange failed — check WATSONX_API_KEY in .env" -ForegroundColor Red
    Write-Host "  Generate a new API key at: https://cloud.ibm.com/iam/apikeys" -ForegroundColor Red
}
Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
