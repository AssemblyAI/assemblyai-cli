Write-Host "Using the OSS distribution..."

$RELEASES_URL = "https://github.com/AssemblyAI/assemblyai-cli/releases"
$FILE_BASENAME = "assemblyai"

$ARCH = $env:PROCESSOR_ARCHITECTURE.ToLower()

$TEMP_DIR = New-TemporaryFile
$TEMP_DIR = $TEMP_DIR.DirectoryName


$RELEASES = Invoke-WebRequest -Uri $RELEASES_URL -UseBasicParsing
$RELEASES = $RELEASES.Links | Where-Object { $_.href -match "download" }
$RELEASES = $RELEASES | Select-Object -Last 1
$RELEASES = $RELEASES.href
$RELEASES = $RELEASES -replace "^.*/", ""
$VERSION = $RELEASES -split "_" | Select-Object -Index 1

Write-Host "Downloading AssemblyAI  $VERSION"

$TAR_FILE = $TEMP_DIR + "/" + $FILE_BASENAME + "_windows_" + $ARCH + ".tar.gz"
$URL = $RELEASES_URL + "/download/v" + $VERSION + "/" + $FILE_BASENAME + "_" + $VERSION + "_windows_" + $ARCH + ".tar.gz"
Write-Host $URL
Start-Process -FilePath "curl" -ArgumentList "-L", $URL, "-o", $TAR_FILE -Wait -NoNewWindow

Write-Host "Extracting AssemblyAI  $VERSION"
Start-Process -FilePath "tar" -ArgumentList @("-xzf", $TAR_FILE, "-C", $TEMP_DIR) -Wait -NoNewWindow
Remove-Item $TAR_FILE

Write-Host "Installing AssemblyAI  $VERSION"
$BINARY = $TEMP_DIR + "/" + $FILE_BASENAME + ".exe"
$INSTALL_DIR = $env:ProgramFiles + "/AssemblyAI"
if (!(Test-Path $INSTALL_DIR)) {
  New-Item -ItemType Directory -Path $INSTALL_DIR
}
$TARGET = $INSTALL_DIR + "/" + $FILE_BASENAME + ".exe"
Copy-Item $BINARY $TARGET -Force

Start-Process -FilePath $TARGET -ArgumentList "welcome -i -o=windows -m=curl -v=$VERSION -a=$ARCH" -Wait -NoNewWindow