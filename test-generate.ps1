# test-generate.ps1  v2
# Provides full diagnostic output — server logs will show the real error.
# Run from workspace root with server running: .\test-generate.ps1

$uri = "http://localhost:8080/api/generate"

$json = @{
    session_id = ""
    intake     = @{
        core_idea       = "eco-friendly coffee brand"
        target_audience = "young professionals 25-35"
        personality     = "minimal, warm, honest"
        color_mood      = "green, cream, terracotta"
    }
    action = "init"
} | ConvertTo-Json -Depth 5

Write-Host ">>> POST $uri" -ForegroundColor White
Write-Host $json
Write-Host ""

try {
    $response = Invoke-WebRequest `
        -Uri         $uri `
        -Method      POST `
        -Body        $json `
        -ContentType "application/json" `
        -UseBasicParsing `
        -TimeoutSec  75

    Write-Host "=== HTTP $($response.StatusCode) ===" -ForegroundColor Green
    $data = $response.Content | ConvertFrom-Json
    Write-Host "Session ID : $($data.session_id)"
    Write-Host "Cards      : $($data.cards.Count)"
    if ($data.cards.Count -gt 0) {
        foreach ($c in $data.cards) {
            Write-Host "  >> $($c.name) — $($c.tagline)" -ForegroundColor Cyan
        }
    }
} catch [System.Net.WebException] {
    $ex       = $_.Exception
    $httpResp = $ex.Response
    Write-Host "=== HTTP ERROR ===" -ForegroundColor Red
    if ($httpResp) {
        $statusCode = [int]$httpResp.StatusCode
        Write-Host "Status: $statusCode"
        $stream = $httpResp.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $body   = $reader.ReadToEnd()
        Write-Host "Body  : $body" -ForegroundColor Yellow
    } else {
        Write-Host "No HTTP response — server may not be running or timed out"
        Write-Host $ex.Message
    }
} catch {
    Write-Host "=== UNEXPECTED ERROR ===" -ForegroundColor Red
    Write-Host $_.Exception.Message
}

Write-Host ""
Write-Host ">>> Also check the server terminal for Go log output" -ForegroundColor DarkGray
