# 🧮 Calculette Comptable Pro

Une calculatrice portable spécialisée pour la comptabilité, développée en Go avec interface graphique Fyne.

![Version](https://img.shields.io/badge/version-1.0.0-green)
![Platform](https://img.shields.io/badge/platform-Windows-blue)
![License](https://img.shields.io/badge/license-MIT-yellow)

## ✨ Fonctionnalités

### Fonctions de base
- ➕ Addition, soustraction, multiplication, division
- 🔢 Grand écran avec historique des opérations
- 💾 Mémoire (MC, MR, M+, M-, MS)

### Fonctions comptables
- 📊 **Calcul TVA** : 20%, 10%, 5.5%, 2.1% (taux français)
- 🔄 **Conversion HT ↔ TTC** : En un clic
- 📈 **Pourcentages** : Calculs automatiques
- ± **Changement de signe**
- 📋 **Historique** : Gardez trace de tous vos calculs

## 🚀 Installation Rapide

### Option 1 : Télécharger l'exécutable
Si vous avez reçu le fichier `calculette-comptable.exe`, il suffit de :
1. Copier le fichier sur votre clé USB ou PC
2. Double-cliquer pour lancer - **Aucune installation requise !**

### Option 2 : Compiler depuis les sources

#### Prérequis
1. **Installer Go** : https://go.dev/dl/
   - Téléchargez et installez Go pour Windows
   - Vérifiez l'installation : `go version`

2. **Installer les outils de compilation C** (nécessaire pour Fyne) :
   - Installez [MSYS2](https://www.msys2.org/)
   - Ou installez [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)

#### Compilation

```powershell
# Naviguer vers le dossier du projet
cd "C:\Users\frevi\Desktop\Mes projets\Calculette"

# Télécharger les dépendances
go mod tidy

# Compiler en exécutable portable (sans console)
go build -ldflags="-H windowsgui -s -w" -o calculette-comptable.exe
```

#### Script automatique
Exécutez simplement le script PowerShell fourni :
```powershell
.\build.ps1
```

## 📁 Structure du projet

```
Calculette/
├── go.mod              # Dépendances Go
├── main.go             # Code source principal
├── build.ps1           # Script de compilation
├── README.md           # Ce fichier
└── calculette-comptable.exe  # Exécutable (après compilation)
```

## 🎨 Personnalisation

### Modifier les couleurs

Ouvrez `main.go` et modifiez les variables au début du fichier :

```go
// CONFIGURATION DU STYLE - PERSONNALISABLE
var (
    CouleurPrimaire    = "#1B4D3E" // Vert foncé comptable
    CouleurSecondaire  = "#2E7D52" // Vert moyen
    CouleurAccent      = "#4CAF50" // Vert clair pour les accents
    CouleurTexte       = "#FFFFFF" // Blanc
    CouleurFond        = "#0D2818" // Fond très sombre
    CouleurEcran       = "#1A1A2E" // Fond de l'écran LCD
    CouleurResultat    = "#00FF88" // Couleur du résultat
)
```

### Modifier les taux de TVA

```go
// Taux de TVA par défaut (personnalisable)
var (
    TauxTVAStandard = 20.0  // TVA standard France
    TauxTVAReduit   = 10.0  // TVA réduite
    TauxTVAReduit2  = 5.5   // TVA réduite 2
    TauxTVASuper    = 2.1   // TVA super réduite
)
```

### Ajouter de nouvelles fonctions

Pour ajouter une nouvelle fonction comptable :

1. Créez la méthode dans la section "FONCTIONS COMPTABLES" :
```go
func (c *Calculatrice) maNouvelleFonction() {
    valeur := c.obtenirValeurCourante()
    // Votre logique ici
    resultat := valeur * 2  // exemple
    
    c.valeurCourante = fmt.Sprintf("%.10f", resultat)
    c.resultatAffiche = true
    c.mettreAJourAffichage()
}
```

2. Ajoutez un bouton dans `construireInterface()` :
```go
c.boutonFonction("Ma Fonction", c.maNouvelleFonction),
```

## 🔧 Compilation avancée

### Réduire la taille de l'exe

```powershell
# Compilation optimisée (supprime symboles de debug)
go build -ldflags="-H windowsgui -s -w" -o calculette-comptable.exe

# Compression avec UPX (optionnel, réduit ~60%)
upx --best calculette-comptable.exe
```

### Ajouter une icône

1. Installez fyne-cross ou go-winres
2. Créez un fichier `icon.ico`
3. Utilisez :
```powershell
go install github.com/tc-hib/go-winres@latest
go-winres make --icon icon.ico
go build -ldflags="-H windowsgui -s -w" -o calculette-comptable.exe
```

## 📝 Utilisation

| Bouton | Fonction |
|--------|----------|
| `C` | Effacer tout |
| `CE` | Effacer entrée courante |
| `⌫` | Retour arrière |
| `MC` | Effacer mémoire |
| `MR` | Rappeler mémoire |
| `M+` | Ajouter à mémoire |
| `M-` | Soustraire de mémoire |
| `MS` | Stocker en mémoire |
| `TVA X%` | Calculer la TVA sur le montant affiché |
| `HT→TTC` | Convertir HT en TTC (TVA 20%) |
| `TTC→HT` | Convertir TTC en HT (TVA 20%) |
| `%` | Pourcentage |
| `±` | Changer le signe |

## 🐛 Problèmes connus

- **Le fichier exe est volumineux (~15-30 Mo)** : C'est normal, il contient tout le runtime Go et les bibliothèques graphiques. Vous pouvez le réduire avec UPX.
- **Erreur "gcc not found"** : Installez un compilateur C (TDM-GCC ou MSYS2).

## 📄 Licence

MIT - Libre d'utilisation et de modification.

---

Développé avec ❤️ en Go + Fyne

