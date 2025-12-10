# ============================================
# Script de compilation - Calculette Comptable
# ============================================

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Compilation Calculette Comptable Pro" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Vérifier que Go est installé
try {
    $goVersion = go version
    Write-Host "[OK] Go trouvé: $goVersion" -ForegroundColor Green
}
catch {
    Write-Host "[ERREUR] Go n'est pas installé ou pas dans le PATH" -ForegroundColor Red
    Write-Host "Téléchargez Go sur: https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}

# Télécharger les dépendances
Write-Host ""
Write-Host "[...] Téléchargement des dépendances..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERREUR] Échec du téléchargement des dépendances" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Dépendances téléchargées" -ForegroundColor Green

# Compiler l'application
Write-Host ""
Write-Host "[...] Compilation de l'exécutable..." -ForegroundColor Yellow

$env:CGO_ENABLED = "1"
go build -ldflags="-H windowsgui -s -w" -o calculette-comptable.exe

if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERREUR] Échec de la compilation" -ForegroundColor Red
    Write-Host ""
    Write-Host "Assurez-vous d'avoir installé un compilateur C:" -ForegroundColor Yellow
    Write-Host "  - Option 1: TDM-GCC (https://jmeubank.github.io/tdm-gcc/)" -ForegroundColor Yellow
    Write-Host "  - Option 2: MSYS2 (https://www.msys2.org/)" -ForegroundColor Yellow
    exit 1
}

# Afficher les informations sur le fichier
$exeFile = Get-Item "calculette-comptable.exe"
$sizeMB = [math]::Round($exeFile.Length / 1MB, 2)

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  COMPILATION RÉUSSIE !" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Fichier créé: calculette-comptable.exe" -ForegroundColor White
Write-Host "Taille: $sizeMB Mo" -ForegroundColor White
Write-Host "Emplacement: $($exeFile.FullName)" -ForegroundColor White
Write-Host ""
Write-Host "L'exécutable est portable et peut être copié" -ForegroundColor Cyan
Write-Host "sur n'importe quel PC Windows (clé USB, etc.)" -ForegroundColor Cyan
Write-Host ""

# Proposer de lancer l'application
$response = Read-Host "Voulez-vous lancer l'application maintenant ? (O/N)"
if ($response -eq "O" -or $response -eq "o" -or $response -eq "oui") {
    Write-Host "Lancement de l'application..." -ForegroundColor Green
    Start-Process ".\calculette-comptable.exe"
}

