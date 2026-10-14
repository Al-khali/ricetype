# `ricetype` — `/dev/tty` Typing Trainer for Unix & Rice Enjoyers

```text
       _          _                    
  _ __(_) ___ ___| |_ _   _ _ __   ___ 
 | '__| |/ __/ _ \ __| | | | '_ \ / _ \
 | |  | | (_|  __/ |_| |_| | |_) |  __/
 |_|  |_|\___\___|\__|\__, | .__/ \___|
                      |___/|_|         
```

**`ricetype`** est un entraîneur de dactylographie TUI minimaliste, ultra-rapide (`0 allocs/op` par frappe) et pensé pour s'intégrer parfaitement dans votre setup **Unix / Hyprland / Sway / Neovim / tmux** ricé.

Construit en **Go** avec [Bubble Tea](https://github.com/charmbracelet/bubbletea) et [Lipgloss](https://github.com/charmbracelet/lipgloss), il combine l'ergonomie des meilleurs tests de frappe modernes avec des palettes natives pour vos dotfiles et une empreinte mémoire minuscule.

---

## Pourquoi `ricetype` ?

- **Esthétique `/r/unixporn` native** : Basculez à la volée (`Ctrl+T`) entre 6 palettes iconiques pré-intégrées :
  - **Catppuccin Mocha**
  - **Tokyo Night**
  - **Dracula**
  - **Nord**
  - **Gruvbox Dark**
  - **Cyberpunk**
- **Zéro Bloat, Zéro Latence (`0 allocs/op`)** :
  - **Run-Length ANSI Batching** : Les caractères adjacents partageant le même style sont fusionnés en une seule séquence d'échappement ANSI au lieu d'être stylisés caractère par caractère.
  - **Wrapping pré-calculé (`LineSpan`)** : Aucun `strings.Split` ni allocation de `lipgloss.NewStyle()` dans la boucle de rendu `View()`.
  - **Moteur incrémental** : Calcul du Net WPM, Raw WPM, Précision, Consistance et Combo en $O(1)$ à chaque frappe.
- **Workflow Clavier Pur (Vim / Unix friendly)** :
  - Navigation au menu en `hjkl` ou flèches, suppression de mot complet via `Ctrl+W` ou `Alt+Backspace`, redémarrage instantané sur `Tab` ou `Ctrl+R`, et **Mode Zen** (`Ctrl+Z`) pour masquer le HUD et ne garder que le flux de texte.

---

## Fonctionnalités

### 4 Modes d'Entraînement
- **`⏱ temps`** (`15s`, `30s`, `60s`, `120s`) : Génération dynamique continue de mots au fil de la frappe.
- **`✎ mots`** (`10`, `25`, `50`, `100` mots) : Course de vitesse sur un nombre fixe de mots.
- **`❝ citation`** : Citations cultes de la philosophie Unix, de l'informatique (Dijkstra, Pike, Torvalds, Knuth, Kay...) et de la littérature.
- **`☠ mort subite`** : La moindre erreur de frappe met immédiatement fin à la session.

### 3 Corpus Embarqués (`//go:embed`) & Modificateurs
- **`FR`** (Français), **`EN`** (English), et **`</> Code`** (expressions idiomatiques Go, Rust, TypeScript, commandes Unix/Git/Docker).
- **`@ ponctuation`** & **`# nombres`** : Activation instantanée depuis le menu (`P` / `N`) pour injecter majuscules, signes de ponctuation, parenthèses et nombres.

### Télémétrie, Graphique Unicode & Records Locaux
- **Viewport Glissant 3 Lignes** avec barre de progression Unicode fluide (`━━━╸───────`).
- **Écran de Fin Détaillé** :
  - **Net WPM**, **Raw WPM**, **Précision (%)**, **Consistance (%)** (calculée sur l'écart-type seconde par seconde), détail des frappes (`correctes / erreurs / corrigées`) et **Combo Max**.
  - **Courbe WPM Unicode graduée** avec axe vertical et marqueurs d'erreurs temporels (`×`).
  - **Badges de Rangs** (*Novice Tranquille* $\rightarrow$ *Cyber-Dieu du Clavier*) et animation de célébration lors d'un **Nouveau Record Personnel (PB)**.
- **Persistance XDG Config** : Sauvegarde automatique de vos records par catégorie, de votre historique et de votre thème préféré dans `~/.config/ricetype/stats.json`.

---

## Raccourcis Clavier

| Raccourci | Action |
| :--- | :--- |
| `h / j / k / l` ou `← / ↓ / ↑ / →` | Naviguer et ajuster les pilules de configuration |
| `Entrée` ou `Espace` | Lancer le test (ou basculer un modificateur sur la ligne Options) |
| `P` / `N` *(Menu)* | Activer/désactiver `@ ponctuation` (`P`) ou `# nombres` (`N`) |
| `S` *(Menu / Résultats)* | Afficher les Records Personnels (PB) & l'historique des sessions |
| `Ctrl + W` / `Alt + Backspace` | Supprimer le mot précédent entier pendant la frappe |
| `Tab` / `Ctrl + R` | Relancer immédiatement un test avec un nouveau tirage |
| `Ctrl + Z` | Activer/désactiver le **Mode Zen** (masquer le HUD live) |
| `Ctrl + T` | Changer de thème de couleurs à la volée (sauvegardé automatiquement) |
| `Esc` | Retourner au menu principal |
| `Q` / `Ctrl + C` | Quitter proprement |

---

## Installation & Utilisation

### Prérequis
- **Go 1.22+**

### Cloner et compiler
```bash
git clone https://github.com/Al-khali/ricetype.git
cd ricetype
go build -ldflags="-s -w" -o ricetype .
./ricetype
```

### Lancer directement
```bash
go run .
```

### Lancer les tests et le benchmark mémoire
```bash
go test -v -bench=. -benchmem ./internal/...
```

```text
BenchmarkRewrapAndType-10    1000000    4184 ns/op    66 B/op    0 allocs/op
```

---

## Architecture du Projet

```text
ricetype/
├── main.go                     # Point d'entrée Bubble Tea (AltScreen)
└── internal/
    ├── engine/                 # Moteur de frappe O(1), wrapping LineSpan & stats
    ├── words/                  # Corpus embarqués (go:embed FR/EN/Code/Quotes)
    ├── theme/                  # Palettes r/unixporn & cache de styles Lipgloss
    ├── storage/                # Persistance JSON (~/.config/ricetype/stats.json)
    └── ui/                     # Modèle Bubble Tea, Viewport 3 lignes & Batching ANSI
```
