package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ========================================
// CONFIGURATION DU STYLE - PERSONNALISABLE
// ========================================

// Couleurs principales (format hexadécimal)
var (
	CouleurPrimaire   = "#1B4D3E" // Vert foncé comptable
	CouleurSecondaire = "#2E7D52" // Vert moyen
	CouleurAccent     = "#4CAF50" // Vert clair pour les accents
	CouleurTexte      = "#FFFFFF" // Blanc
	CouleurFond       = "#0D2818" // Fond très sombre
	CouleurEcran      = "#1A1A2E" // Fond de l'écran LCD
	CouleurResultat   = "#00FF88" // Couleur du résultat (vert fluo)
)

// Taux de TVA par défaut (personnalisable)
var (
	TauxTVAStandard = 20.0 // TVA standard France
	TauxTVAReduit   = 10.0 // TVA réduite
	TauxTVAReduit2  = 5.5  // TVA réduite 2
	TauxTVASuper    = 2.1  // TVA super réduite
)

// ========================================
// STRUCTURE DE L'APPLICATION
// ========================================

type Calculatrice struct {
	affichage       *widget.Label
	sousAffichage   *widget.Label
	historique      *widget.List
	listeHistorique []string

	valeurCourante   string
	valeurPrecedente string
	operation        string
	resultatAffiche  bool
	memoireM         float64
	fenetre          fyne.Window
}

func main() {
	// Création de l'application
	a := app.New()
	a.Settings().SetTheme(&monTheme{})

	// Fenêtre principale
	w := a.NewWindow("Calculette Comptable Pro")
	w.Resize(fyne.NewSize(700, 750))
	w.SetIcon(creerIcone())

	// Instance de la calculatrice
	calc := &Calculatrice{
		listeHistorique: make([]string, 0),
		fenetre:         w,
	}

	// Construction de l'interface
	contenu := calc.construireInterface()
	w.SetContent(contenu)

	// === RACCOURCIS GLOBAUX (Ctrl+C, Ctrl+V) ===
	w.Canvas().AddShortcut(&fyne.ShortcutCopy{}, func(shortcut fyne.Shortcut) {
		calc.copierVersClipboard()
	})

	w.Canvas().AddShortcut(&fyne.ShortcutPaste{}, func(shortcut fyne.Shortcut) {
		calc.collerDepuisClipboard()
	})

	// === GESTION DU CLAVIER GLOBAL ===
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		switch k.Name {
		case fyne.KeyReturn, fyne.KeyEnter:
			calc.calculer()
		case fyne.KeyBackspace:
			calc.retourArriere()
		case fyne.KeyEscape:
			calc.effacerTout()
		case fyne.KeyDelete:
			calc.effacerCourant()
		}
	})

	w.Canvas().SetOnTypedRune(func(r rune) {
		switch r {
		// Chiffres
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			calc.ajouterChiffre(string(r))
		// Opérations
		case '+':
			calc.definirOperation("+")
		case '-':
			calc.definirOperation("-")
		case '*', 'x', 'X':
			calc.definirOperation("*")
		case '/', ':':
			calc.definirOperation("/")
		// Égal
		case '=':
			calc.calculer()
		// Virgule / Point décimal
		case ',', '.':
			calc.ajouterVirgule()
		// Pourcentage
		case '%':
			calc.pourcentage()
		}
	})

	w.ShowAndRun()
}

func (c *Calculatrice) construireInterface() fyne.CanvasObject {
	// === ZONE D'AFFICHAGE ===
	c.sousAffichage = widget.NewLabel("")
	c.sousAffichage.Alignment = fyne.TextAlignTrailing

	c.affichage = widget.NewLabel("0")
	c.affichage.Alignment = fyne.TextAlignTrailing
	c.affichage.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	// Conteneur d'affichage avec fond
	ecranFond := canvas.NewRectangle(hexToCouleur(CouleurEcran))
	ecranFond.CornerRadius = 8

	// Indication pour copier
	indiceCopie := widget.NewLabel("Double-clic pour copier")
	indiceCopie.Alignment = fyne.TextAlignLeading
	indiceCopie.TextStyle = fyne.TextStyle{Italic: true}

	ecranContenu := container.NewVBox(
		indiceCopie,
		c.sousAffichage,
		c.affichage,
	)

	// Écran cliquable (double-clic = copier)
	ecranCliquable := newEcranTappable(c)
	ecran := container.NewStack(ecranFond, container.NewPadded(ecranContenu), ecranCliquable)

	// === BOUTONS MÉMOIRE ===
	btnsMem := container.NewGridWithColumns(5,
		c.boutonMem("MC", c.memClear),
		c.boutonMem("MR", c.memRecall),
		c.boutonMem("M+", c.memAdd),
		c.boutonMem("M-", c.memSub),
		c.boutonMem("MS", c.memStore),
	)

	// === BOUTONS TVA (COMPTABILITÉ) ===
	btnsTVA := container.NewGridWithColumns(4,
		c.boutonTVA(fmt.Sprintf("%.0f%%", TauxTVAStandard), TauxTVAStandard),
		c.boutonTVA(fmt.Sprintf("%.0f%%", TauxTVAReduit), TauxTVAReduit),
		c.boutonTVA(fmt.Sprintf("%.1f%%", TauxTVAReduit2), TauxTVAReduit2),
		c.boutonTVA(fmt.Sprintf("%.1f%%", TauxTVASuper), TauxTVASuper),
	)

	// === BOUTONS FONCTIONS COMPTABLES ===
	btnsCompta := container.NewGridWithColumns(4,
		c.boutonFonction("HT>TTC", c.htVersTTC),
		c.boutonFonction("TTC>HT", c.ttcVersHT),
		c.boutonFonction("%", c.pourcentage),
		c.boutonFonction("+/-", c.changerSigne),
	)

	// === PAVÉ NUMÉRIQUE PRINCIPAL ===
	paveNum := container.NewGridWithColumns(4,
		// Ligne 1
		c.boutonEffacer("C", c.effacerTout),
		c.boutonEffacer("CE", c.effacerCourant),
		c.boutonEffacer("<-", c.retourArriere),
		c.boutonOperation("/", "/"),
		// Ligne 2
		c.boutonChiffre("7"),
		c.boutonChiffre("8"),
		c.boutonChiffre("9"),
		c.boutonOperation("x", "*"),
		// Ligne 3
		c.boutonChiffre("4"),
		c.boutonChiffre("5"),
		c.boutonChiffre("6"),
		c.boutonOperation("-", "-"),
		// Ligne 4
		c.boutonChiffre("1"),
		c.boutonChiffre("2"),
		c.boutonChiffre("3"),
		c.boutonOperation("+", "+"),
		// Ligne 5
		c.boutonChiffre("00"),
		c.boutonChiffre("0"),
		c.boutonVirgule(),
		c.boutonEgal(),
	)

	// === HISTORIQUE ===
	c.historique = widget.NewList(
		func() int { return len(c.listeHistorique) },
		func() fyne.CanvasObject {
			return widget.NewLabel("                              ")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(c.listeHistorique[len(c.listeHistorique)-1-i])
		},
	)

	// Clic sur une ligne d'historique = charger le résultat
	c.historique.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(c.listeHistorique) {
			ligne := c.listeHistorique[len(c.listeHistorique)-1-id]
			c.chargerDepuisHistorique(ligne)
		}
		// Désélectionner visuellement
		c.historique.UnselectAll()
	}

	titreHisto := widget.NewLabel("Historique")
	titreHisto.TextStyle = fyne.TextStyle{Bold: true}

	btnEffHisto := widget.NewButton("Effacer", func() {
		c.listeHistorique = make([]string, 0)
		c.historique.Refresh()
	})

	headerHisto := container.NewBorder(nil, nil, titreHisto, btnEffHisto)
	sectionHisto := container.NewBorder(headerHisto, nil, nil, nil,
		container.NewVScroll(c.historique))

	// === ASSEMBLAGE FINAL ===
	partieCalcul := container.NewVBox(
		ecran,
		widget.NewSeparator(),
		btnsMem,
		widget.NewSeparator(),
		btnsTVA,
		btnsCompta,
		widget.NewSeparator(),
		paveNum,
	)

	// Diviseur vertical : calculatrice | historique
	split := container.NewHSplit(
		container.NewPadded(partieCalcul),
		container.NewPadded(sectionHisto),
	)
	split.SetOffset(0.65)

	return split
}

// ========================================
// CRÉATION DES BOUTONS
// ========================================

func (c *Calculatrice) boutonChiffre(chiffre string) *widget.Button {
	btn := widget.NewButton(chiffre, func() {
		c.ajouterChiffre(chiffre)
	})
	btn.Importance = widget.MediumImportance
	return btn
}

func (c *Calculatrice) boutonOperation(label, op string) *widget.Button {
	btn := widget.NewButton(label, func() {
		c.definirOperation(op)
	})
	btn.Importance = widget.HighImportance
	return btn
}

func (c *Calculatrice) boutonEffacer(label string, action func()) *widget.Button {
	btn := widget.NewButton(label, action)
	btn.Importance = widget.DangerImportance
	return btn
}

func (c *Calculatrice) boutonVirgule() *widget.Button {
	btn := widget.NewButton(",", func() {
		c.ajouterVirgule()
	})
	btn.Importance = widget.MediumImportance
	return btn
}

func (c *Calculatrice) boutonEgal() *widget.Button {
	btn := widget.NewButton("=", func() {
		c.calculer()
	})
	btn.Importance = widget.SuccessImportance
	return btn
}

func (c *Calculatrice) boutonMem(label string, action func()) *widget.Button {
	btn := widget.NewButton(label, action)
	btn.Importance = widget.LowImportance
	return btn
}

func (c *Calculatrice) boutonTVA(label string, taux float64) *widget.Button {
	btn := widget.NewButton(label, func() {
		c.calculerTVA(taux)
	})
	btn.Importance = widget.WarningImportance
	return btn
}

func (c *Calculatrice) boutonFonction(label string, action func()) *widget.Button {
	btn := widget.NewButton(label, action)
	btn.Importance = widget.HighImportance
	return btn
}

// ========================================
// LOGIQUE DE CALCUL
// ========================================

func (c *Calculatrice) ajouterChiffre(chiffre string) {
	if c.resultatAffiche {
		c.valeurCourante = ""
		c.resultatAffiche = false
	}

	// Limiter la longueur
	if len(c.valeurCourante) >= 15 {
		return
	}

	c.valeurCourante += chiffre
	c.mettreAJourAffichage()
}

func (c *Calculatrice) ajouterVirgule() {
	if c.resultatAffiche {
		c.valeurCourante = "0"
		c.resultatAffiche = false
	}

	if c.valeurCourante == "" {
		c.valeurCourante = "0"
	}

	if !strings.Contains(c.valeurCourante, ".") {
		c.valeurCourante += "."
	}
	c.mettreAJourAffichage()
}

func (c *Calculatrice) definirOperation(op string) {
	if c.valeurCourante == "" && c.valeurPrecedente == "" {
		return
	}

	if c.valeurPrecedente != "" && c.valeurCourante != "" {
		c.calculer()
	}

	if c.valeurCourante != "" {
		c.valeurPrecedente = c.valeurCourante
	}
	c.operation = op
	c.valeurCourante = ""
	c.resultatAffiche = false

	// Afficher l'opération en cours
	c.sousAffichage.SetText(fmt.Sprintf("%s %s", c.formaterNombre(c.valeurPrecedente), c.symbolOperation()))
}

func (c *Calculatrice) symbolOperation() string {
	switch c.operation {
	case "+":
		return "+"
	case "-":
		return "-"
	case "*":
		return "x"
	case "/":
		return "/"
	default:
		return ""
	}
}

func (c *Calculatrice) calculer() {
	if c.valeurPrecedente == "" || c.valeurCourante == "" || c.operation == "" {
		return
	}

	a, _ := strconv.ParseFloat(c.valeurPrecedente, 64)
	b, _ := strconv.ParseFloat(c.valeurCourante, 64)

	var resultat float64
	var expression string

	switch c.operation {
	case "+":
		resultat = a + b
		expression = fmt.Sprintf("%s + %s", c.formaterNombre(c.valeurPrecedente), c.formaterNombre(c.valeurCourante))
	case "-":
		resultat = a - b
		expression = fmt.Sprintf("%s - %s", c.formaterNombre(c.valeurPrecedente), c.formaterNombre(c.valeurCourante))
	case "*":
		resultat = a * b
		expression = fmt.Sprintf("%s x %s", c.formaterNombre(c.valeurPrecedente), c.formaterNombre(c.valeurCourante))
	case "/":
		if b == 0 {
			c.affichage.SetText("Erreur: /0")
			c.sousAffichage.SetText("")
			c.reinitialiser()
			return
		}
		resultat = a / b
		expression = fmt.Sprintf("%s / %s", c.formaterNombre(c.valeurPrecedente), c.formaterNombre(c.valeurCourante))
	}

	// Ajouter à l'historique
	resultatStr := c.formaterResultat(resultat)
	c.ajouterHistorique(fmt.Sprintf("%s = %s", expression, resultatStr))

	c.valeurCourante = fmt.Sprintf("%.10f", resultat)
	c.valeurCourante = strings.TrimRight(strings.TrimRight(c.valeurCourante, "0"), ".")
	c.valeurPrecedente = ""
	c.operation = ""
	c.resultatAffiche = true

	c.sousAffichage.SetText(expression + " =")
	c.affichage.SetText(resultatStr)
}

func (c *Calculatrice) effacerTout() {
	c.reinitialiser()
	c.sousAffichage.SetText("")
	c.affichage.SetText("0")
}

func (c *Calculatrice) effacerCourant() {
	c.valeurCourante = ""
	c.affichage.SetText("0")
}

func (c *Calculatrice) retourArriere() {
	if c.resultatAffiche || c.valeurCourante == "" {
		return
	}
	c.valeurCourante = c.valeurCourante[:len(c.valeurCourante)-1]
	c.mettreAJourAffichage()
}

func (c *Calculatrice) reinitialiser() {
	c.valeurCourante = ""
	c.valeurPrecedente = ""
	c.operation = ""
	c.resultatAffiche = false
}

// ========================================
// FONCTIONS COMPTABLES
// ========================================

func (c *Calculatrice) calculerTVA(taux float64) {
	valeur := c.obtenirValeurCourante()
	if valeur == 0 {
		return
	}

	tva := valeur * taux / 100
	expression := fmt.Sprintf("TVA %.1f%% de %s", taux, c.formaterResultat(valeur))
	resultat := c.formaterResultat(tva)

	c.valeurCourante = fmt.Sprintf("%.10f", tva)
	c.valeurCourante = strings.TrimRight(strings.TrimRight(c.valeurCourante, "0"), ".")
	c.resultatAffiche = true

	c.sousAffichage.SetText(expression)
	c.affichage.SetText(resultat)
	c.ajouterHistorique(fmt.Sprintf("%s = %s", expression, resultat))
}

func (c *Calculatrice) htVersTTC() {
	valeur := c.obtenirValeurCourante()
	if valeur == 0 {
		return
	}

	ttc := valeur * (1 + TauxTVAStandard/100)
	expression := fmt.Sprintf("%s HT > TTC (%.0f%%)", c.formaterResultat(valeur), TauxTVAStandard)
	resultat := c.formaterResultat(ttc)

	c.valeurCourante = fmt.Sprintf("%.10f", ttc)
	c.valeurCourante = strings.TrimRight(strings.TrimRight(c.valeurCourante, "0"), ".")
	c.resultatAffiche = true

	c.sousAffichage.SetText(expression)
	c.affichage.SetText(resultat)
	c.ajouterHistorique(fmt.Sprintf("%s = %s", expression, resultat))
}

func (c *Calculatrice) ttcVersHT() {
	valeur := c.obtenirValeurCourante()
	if valeur == 0 {
		return
	}

	ht := valeur / (1 + TauxTVAStandard/100)
	expression := fmt.Sprintf("%s TTC > HT (%.0f%%)", c.formaterResultat(valeur), TauxTVAStandard)
	resultat := c.formaterResultat(ht)

	c.valeurCourante = fmt.Sprintf("%.10f", ht)
	c.valeurCourante = strings.TrimRight(strings.TrimRight(c.valeurCourante, "0"), ".")
	c.resultatAffiche = true

	c.sousAffichage.SetText(expression)
	c.affichage.SetText(resultat)
	c.ajouterHistorique(fmt.Sprintf("%s = %s", expression, resultat))
}

func (c *Calculatrice) pourcentage() {
	if c.valeurPrecedente != "" && c.valeurCourante != "" {
		// Calcul de pourcentage : X + Y% ou X - Y%
		base, _ := strconv.ParseFloat(c.valeurPrecedente, 64)
		pourcent, _ := strconv.ParseFloat(c.valeurCourante, 64)
		resultat := base * pourcent / 100

		c.valeurCourante = fmt.Sprintf("%.10f", resultat)
		c.valeurCourante = strings.TrimRight(strings.TrimRight(c.valeurCourante, "0"), ".")
		c.mettreAJourAffichage()
	} else if c.valeurCourante != "" {
		// Simple conversion en pourcentage
		valeur := c.obtenirValeurCourante()
		resultat := valeur / 100

		expression := fmt.Sprintf("%s%%", c.formaterResultat(valeur))
		resultatStr := c.formaterResultat(resultat)

		c.valeurCourante = fmt.Sprintf("%.10f", resultat)
		c.valeurCourante = strings.TrimRight(strings.TrimRight(c.valeurCourante, "0"), ".")
		c.resultatAffiche = true

		c.sousAffichage.SetText(expression)
		c.affichage.SetText(resultatStr)
	}
}

func (c *Calculatrice) changerSigne() {
	if c.valeurCourante == "" || c.valeurCourante == "0" {
		return
	}

	if strings.HasPrefix(c.valeurCourante, "-") {
		c.valeurCourante = c.valeurCourante[1:]
	} else {
		c.valeurCourante = "-" + c.valeurCourante
	}
	c.mettreAJourAffichage()
}

// ========================================
// FONCTIONS MÉMOIRE
// ========================================

func (c *Calculatrice) memClear() {
	c.memoireM = 0
}

func (c *Calculatrice) memRecall() {
	c.valeurCourante = fmt.Sprintf("%.10f", c.memoireM)
	c.valeurCourante = strings.TrimRight(strings.TrimRight(c.valeurCourante, "0"), ".")
	c.mettreAJourAffichage()
}

func (c *Calculatrice) memAdd() {
	c.memoireM += c.obtenirValeurCourante()
}

func (c *Calculatrice) memSub() {
	c.memoireM -= c.obtenirValeurCourante()
}

func (c *Calculatrice) memStore() {
	c.memoireM = c.obtenirValeurCourante()
}

// ========================================
// UTILITAIRES
// ========================================

func (c *Calculatrice) obtenirValeurCourante() float64 {
	if c.valeurCourante == "" {
		return 0
	}
	val, _ := strconv.ParseFloat(c.valeurCourante, 64)
	return val
}

func (c *Calculatrice) mettreAJourAffichage() {
	if c.valeurCourante == "" {
		c.affichage.SetText("0")
	} else {
		c.affichage.SetText(c.formaterNombre(c.valeurCourante))
	}
}

func (c *Calculatrice) formaterNombre(s string) string {
	if s == "" {
		return "0"
	}
	// Remplacer le point par une virgule pour l'affichage français
	return strings.ReplaceAll(s, ".", ",")
}

func (c *Calculatrice) formaterResultat(n float64) string {
	// Formater avec 2 décimales pour les montants, sinon intelligent
	if n == float64(int64(n)) {
		return fmt.Sprintf("%.0f", n)
	}

	resultat := fmt.Sprintf("%.2f", n)
	return strings.ReplaceAll(resultat, ".", ",")
}

func (c *Calculatrice) ajouterHistorique(entree string) {
	c.listeHistorique = append(c.listeHistorique, entree)
	c.historique.Refresh()
}

// ========================================
// CLIPBOARD ET HISTORIQUE
// ========================================

// Copier la valeur affichée dans le presse-papier
func (c *Calculatrice) copierVersClipboard() {
	texte := c.affichage.Text
	// Convertir la virgule en point pour le presse-papier
	texte = strings.ReplaceAll(texte, ",", ".")
	c.fenetre.Clipboard().SetContent(texte)
}

// Coller depuis le presse-papier
func (c *Calculatrice) collerDepuisClipboard() {
	contenu := c.fenetre.Clipboard().Content()
	if contenu == "" {
		return
	}

	// Nettoyer le contenu collé
	contenu = strings.TrimSpace(contenu)
	contenu = strings.ReplaceAll(contenu, ",", ".")
	contenu = strings.ReplaceAll(contenu, " ", "")

	// Vérifier que c'est un nombre valide
	_, err := strconv.ParseFloat(contenu, 64)
	if err != nil {
		return
	}

	c.valeurCourante = contenu
	c.resultatAffiche = false
	c.mettreAJourAffichage()
}

// Charger le résultat d'une ligne d'historique
func (c *Calculatrice) chargerDepuisHistorique(ligne string) {
	// Format attendu: "expression = resultat"
	parties := strings.Split(ligne, "=")
	if len(parties) < 2 {
		return
	}

	resultat := strings.TrimSpace(parties[len(parties)-1])
	// Convertir virgule en point pour le stockage interne
	resultat = strings.ReplaceAll(resultat, ",", ".")

	// Vérifier que c'est un nombre valide
	_, err := strconv.ParseFloat(resultat, 64)
	if err != nil {
		return
	}

	c.valeurCourante = resultat
	c.valeurPrecedente = ""
	c.operation = ""
	c.resultatAffiche = true
	c.sousAffichage.SetText("Depuis historique:")
	c.mettreAJourAffichage()
}

// ========================================
// WIDGET ÉCRAN CLIQUABLE
// ========================================

// ecranTappable permet de détecter les double-clics sur l'écran
type ecranTappable struct {
	widget.BaseWidget
	calc *Calculatrice
}

func newEcranTappable(calc *Calculatrice) *ecranTappable {
	e := &ecranTappable{calc: calc}
	e.ExtendBaseWidget(e)
	return e
}

func (e *ecranTappable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}

func (e *ecranTappable) Tapped(_ *fyne.PointEvent) {
	// Simple clic - ne rien faire
}

func (e *ecranTappable) DoubleTapped(_ *fyne.PointEvent) {
	// Double-clic = copier
	e.calc.copierVersClipboard()
	// Feedback visuel
	original := e.calc.sousAffichage.Text
	e.calc.sousAffichage.SetText("Copie!")
	go func() {
		time.Sleep(500 * time.Millisecond)
		e.calc.sousAffichage.SetText(original)
	}()
}

// ========================================
// THÈME PERSONNALISÉ
// ========================================

type monTheme struct{}

func (m *monTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return hexToCouleur(CouleurFond)
	case theme.ColorNameButton:
		return hexToCouleur(CouleurSecondaire)
	case theme.ColorNamePrimary:
		return hexToCouleur(CouleurAccent)
	case theme.ColorNameForeground:
		return hexToCouleur(CouleurTexte)
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m *monTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *monTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *monTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 18 // Police plus grande
	case theme.SizeNameHeadingText:
		return 32 // Titres plus grands
	case theme.SizeNameSubHeadingText:
		return 24
	case theme.SizeNamePadding:
		return 8 // Plus d'espace
	case theme.SizeNameInnerPadding:
		return 12
	case theme.SizeNameScrollBar:
		return 16
	default:
		return theme.DefaultTheme().Size(name) * 1.2 // 20% plus grand par défaut
	}
}

// Conversion couleur hexadécimale vers color.NRGBA
func hexToCouleur(hex string) color.NRGBA {
	hex = strings.TrimPrefix(hex, "#")
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

// ========================================
// ICÔNE DE L'APPLICATION
// ========================================

func creerIcone() fyne.Resource {
	// Créer une icône 64x64 de calculatrice
	taille := 64
	img := image.NewRGBA(image.Rect(0, 0, taille, taille))

	// Couleurs
	vertFonce := color.RGBA{27, 77, 62, 255}  // #1B4D3E
	vertClair := color.RGBA{76, 175, 80, 255} // #4CAF50
	blanc := color.RGBA{255, 255, 255, 255}
	gris := color.RGBA{40, 40, 40, 255}

	// Fond arrondi vert foncé
	for y := 0; y < taille; y++ {
		for x := 0; x < taille; x++ {
			// Coins arrondis (rayon 8)
			dx := min(x, taille-1-x)
			dy := min(y, taille-1-y)
			if dx < 8 && dy < 8 {
				dist := (8-dx)*(8-dx) + (8-dy)*(8-dy)
				if dist > 64 {
					img.Set(x, y, color.RGBA{0, 0, 0, 0})
					continue
				}
			}
			img.Set(x, y, vertFonce)
		}
	}

	// Écran LCD (rectangle gris foncé en haut)
	for y := 8; y < 22; y++ {
		for x := 8; x < 56; x++ {
			img.Set(x, y, gris)
		}
	}

	// Texte "123" sur l'écran
	dessinerChiffre(img, 12, 10, blanc)  // 1
	dessinerChiffre2(img, 24, 10, blanc) // 2
	dessinerChiffre3(img, 36, 10, blanc) // 3

	// Boutons (grille 4x3)
	couleursBoutons := []color.RGBA{vertClair, vertClair, vertClair, vertClair}
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			bx := 8 + col*12
			by := 26 + row*12
			c := couleursBoutons[col%4]
			if col == 3 {
				c = color.RGBA{255, 152, 0, 255} // Orange pour opérations
			}
			for dy := 0; dy < 9; dy++ {
				for dx := 0; dx < 9; dx++ {
					img.Set(bx+dx, by+dy, c)
				}
			}
		}
	}

	// Encoder en PNG
	var buf bytes.Buffer
	png.Encode(&buf, img)

	return fyne.NewStaticResource("icon.png", buf.Bytes())
}

// Dessine le chiffre 1
func dessinerChiffre(img *image.RGBA, x, y int, c color.RGBA) {
	for dy := 0; dy < 9; dy++ {
		img.Set(x+4, y+dy, c)
	}
	img.Set(x+3, y+1, c)
	for dx := 2; dx < 7; dx++ {
		img.Set(x+dx, y+8, c)
	}
}

// Dessine le chiffre 2
func dessinerChiffre2(img *image.RGBA, x, y int, c color.RGBA) {
	for dx := 1; dx < 7; dx++ {
		img.Set(x+dx, y, c)
		img.Set(x+dx, y+4, c)
		img.Set(x+dx, y+8, c)
	}
	img.Set(x+6, y+1, c)
	img.Set(x+6, y+2, c)
	img.Set(x+6, y+3, c)
	img.Set(x+1, y+5, c)
	img.Set(x+1, y+6, c)
	img.Set(x+1, y+7, c)
}

// Dessine le chiffre 3
func dessinerChiffre3(img *image.RGBA, x, y int, c color.RGBA) {
	for dx := 1; dx < 6; dx++ {
		img.Set(x+dx, y, c)
		img.Set(x+dx, y+4, c)
		img.Set(x+dx, y+8, c)
	}
	for dy := 1; dy < 8; dy++ {
		img.Set(x+6, y+dy, c)
	}
}
