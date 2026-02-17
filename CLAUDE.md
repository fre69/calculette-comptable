# Calculette Comptable Pro

Application de bureau Windows - calculatrice orientée comptabilité française.

## Stack technique

- **Langage :** Go 1.21
- **GUI :** Fyne v2.4.3
- **Module :** `calculette-comptable`
- **Architecture :** Application monolithique, fichier unique `main.go` (~900 lignes)
- **Cible :** Windows (portable .exe)

## Compilation

```powershell
# Via le script PowerShell
.\build.ps1

# Ou manuellement
$env:CGO_ENABLED = "1"
go build -ldflags="-H windowsgui -s -w" -o calculette-comptable.exe
```

**Prérequis :** Go 1.21+, compilateur C (TDM-GCC ou MSYS2), CGO activé.

## Structure du code (`main.go`)

Le code est organisé en sections clairement délimitées par des commentaires `// ========` :

| Section | Lignes | Description |
|---------|--------|-------------|
| Configuration style | 21-42 | Couleurs hex, taux TVA (variables globales modifiables) |
| Struct `Calculatrice` | 48-60 | État de l'application (affichage, valeurs, mémoire, historique) |
| `main()` | 62-132 | Point d'entrée, fenêtre Fyne, raccourcis clavier |
| `construireInterface()` | 134-269 | Construction de l'UI (écran, boutons, historique, layout HSplit 65/35) |
| Création boutons | 275-331 | Factories par type (chiffre, opération, effacer, TVA, mémoire) |
| Logique de calcul | 337-473 | Opérations (+,-,*,/), saisie, effacement, état |
| Fonctions comptables | 479-574 | TVA, HT↔TTC, pourcentage, changement signe |
| Fonctions mémoire | 580-600 | MC, MR, M+, M-, MS |
| Utilitaires | 606-643 | Parsing, formatage nombres (virgule FR), historique |
| Clipboard & historique | 650-704 | Copier/coller, chargement depuis historique |
| Widget ecranTappable | 710-740 | Widget custom pour double-clic = copier |
| Thème personnalisé | 746-797 | `monTheme` (couleurs, tailles, 20% plus grand par défaut) |
| Génération icône | 803-904 | Création programmatique d'une icône 64x64 PNG |

## Conventions du projet

- **Langue :** Tout en français (variables, commentaires, labels UI)
- **Séparateur décimal :** virgule à l'affichage, point en interne
- **Formatage résultats :** 2 décimales pour les montants, entier si nombre rond
- **Limite saisie :** 15 caractères max
- **Niveaux d'importance des boutons Fyne :**
  - `MediumImportance` → chiffres
  - `HighImportance` → opérations et fonctions
  - `DangerImportance` → effacement (C, CE, ←)
  - `SuccessImportance` → égal (=)
  - `LowImportance` → mémoire
  - `WarningImportance` → TVA

## Points de personnalisation rapide

- **Couleurs :** Variables globales `Couleur*` (lignes 26-34)
- **Taux TVA :** Variables `TauxTVA*` (lignes 37-42) — HT>TTC et TTC>HT utilisent `TauxTVAStandard` (20%)
- **Taille fenêtre :** `w.Resize(fyne.NewSize(700, 750))` (ligne 69)
- **Taille police :** Méthode `Size()` du thème (lignes 771-788)

## Patterns récurrents

### Ajout d'un nouveau bouton
1. Créer une factory `bouton*()` ou réutiliser une existante
2. L'ajouter dans `construireInterface()` dans le conteneur approprié
3. Lier à une méthode de `Calculatrice`

### Ajout d'une nouvelle fonction comptable
1. Créer une méthode `(c *Calculatrice) maFonction()` dans la section comptable
2. Pattern type : obtenir valeur → calculer → formater → mettre à jour affichage + historique
3. Exemple de pattern (copier `calculerTVA` ou `htVersTTC`)

### Mise à jour de l'affichage après un calcul
```go
c.valeurCourante = fmt.Sprintf("%.10f", resultat)
c.valeurCourante = strings.TrimRight(strings.TrimRight(c.valeurCourante, "0"), ".")
c.resultatAffiche = true
c.sousAffichage.SetText(expression)
c.affichage.SetText(c.formaterResultat(resultat))
c.ajouterHistorique(fmt.Sprintf("%s = %s", expression, resultatStr))
```

## Raccourcis clavier existants

| Touche | Action |
|--------|--------|
| 0-9 | Saisie chiffres |
| +, -, *, x, X, /, : | Opérations |
| = , Entrée | Calculer |
| , ou . | Virgule décimale |
| % | Pourcentage |
| Backspace | Retour arrière |
| Escape | Effacer tout (C) |
| Delete | Effacer courant (CE) |
| Ctrl+C | Copier |
| Ctrl+V | Coller |

## Avertissements

- **Pas de tests unitaires** — en ajouter si le projet grossit
- **Pas de persistence** — l'historique est perdu à la fermeture
- **Single-thread UI** — les goroutines qui touchent l'UI (comme le feedback "Copie!") doivent être prudentes
- L'exe compilé fait ~18 Mo (Fyne + CGO) — compressible avec UPX
