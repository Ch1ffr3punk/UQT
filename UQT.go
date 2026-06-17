package main

import (
	"image/color"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type purpleTheme struct {
	base fyne.Theme
}

func (p *purpleTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(s)
}

func (p *purpleTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 122, G: 110, B: 243, A: 255}
	case theme.ColorNameForegroundOnPrimary:
		return color.White
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 122, G: 110, B: 243, A: 255}
	default:
		return p.base.Color(name, variant)
	}
}

func (p *purpleTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return p.base.Icon(name)
}

func (p *purpleTheme) Size(name fyne.ThemeSizeName) float32 {
	return p.base.Size(name)
}

func foldSubject(subject string) string {
	const maxLen = 78
	if len(subject) <= maxLen {
		return subject
	}

	var result strings.Builder
	remaining := subject

	for len(remaining) > maxLen {
		lastSpace := strings.LastIndex(remaining[:maxLen], " ")
		if lastSpace == -1 {
			lastSpace = maxLen
		}
		result.WriteString(remaining[:lastSpace])
		result.WriteString("\n ")
		remaining = remaining[lastSpace:]
		remaining = strings.TrimPrefix(remaining, " ")
	}
	result.WriteString(remaining)

	return result.String()
}

func removeSignature(lines []string, startIndex int) []string {
	var cleanedLines []string
	signatureFound := false

	for i := startIndex; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "-- ") {
			signatureFound = true
			break
		}

		if !signatureFound {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return cleanedLines
}

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(&purpleTheme{base: theme.DarkTheme()})

	window := myApp.NewWindow("Usenet Quote Tool")
	window.Resize(fyne.NewSize(800, 700))

	textEdit := widget.NewMultiLineEntry()
	textEdit.Wrapping = fyne.TextWrapOff
	textEdit.SetMinRowsVisible(20)
	textEdit.SetPlaceHolder("Paste Usenet article here, then press Quote...")

	isDarkTheme := true
	themeSwitch := widget.NewButton("☀️", nil)

	quoteBtn := widget.NewButton("Quote", func() {
		text := textEdit.Text
		if text == "" {
			return
		}

		lines := strings.Split(text, "\n")

		bodyStart := 0
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				bodyStart = i + 1
				break
			}
		}

		from := "Someone"
		for _, line := range lines {
			if strings.HasPrefix(strings.ToLower(line), "from:") {
				from = strings.TrimSpace(line[5:])
				break
			}
		}

		subject := ""
		for i, line := range lines {
			if strings.HasPrefix(strings.ToLower(line), "subject:") {
				subject = strings.TrimSpace(line[8:])
				for j := i + 1; j < len(lines); j++ {
					nextLine := lines[j]
					if len(nextLine) > 0 && (nextLine[0] == ' ' || nextLine[0] == '\t') {
						subject += " " + strings.TrimSpace(nextLine)
					} else {
						break
					}
				}
				if !strings.HasPrefix(strings.ToLower(subject), "re:") {
					subject = "Re: " + subject
				}
				break
			}
		}

		msgID := ""
		for _, line := range lines {
			if strings.HasPrefix(strings.ToLower(line), "message-id:") {
				msgID = strings.TrimSpace(line[11:])
				break
			}
		}

		newsgroups := ""
		for _, line := range lines {
			if strings.HasPrefix(strings.ToLower(line), "newsgroups:") {
				newsgroups = line
				break
			}
		}

		cleanBodyLines := removeSignature(lines, bodyStart)

		var result strings.Builder

		if subject != "" {
			result.WriteString("Subject: " + foldSubject(subject) + "\n")
		}
		if msgID != "" {
			result.WriteString("References: " + msgID + "\n")
		}
		if newsgroups != "" {
			result.WriteString(newsgroups + "\n")
		}
		if subject != "" || msgID != "" || newsgroups != "" {
			result.WriteString("\n")
		}

		result.WriteString(from + " wrote:\n")

		for _, line := range cleanBodyLines {
			result.WriteString("> " + line + "\n")
		}

		textEdit.SetText(result.String())
	})
	quoteBtn.Importance = widget.HighImportance

	clearBtn := widget.NewButton("Clear", func() {
		textEdit.SetText("")
		window.Clipboard().SetContent("")
	})
	clearBtn.Importance = widget.HighImportance

	infoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		projURL, _ := url.Parse("https://github.com/Ch1ffr3punk/UQT")
		projectLink := widget.NewHyperlink("An Open Source project", projURL)
		okButton := widget.NewButton("OK", func() {
			if overlays := window.Canvas().Overlays(); overlays.Top() != nil {
				overlays.Remove(overlays.Top())
			}
		})
		okButton.Importance = widget.HighImportance
		content := container.NewVBox(
			widget.NewLabelWithStyle("UQT v0.1.0", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			container.NewHBox(layout.NewSpacer(), projectLink, layout.NewSpacer()),
			widget.NewLabelWithStyle("released under the Apache 2.0 license", fyne.TextAlignCenter, fyne.TextStyle{}),
			widget.NewLabelWithStyle("© 2026 Ch1ffr3punk", fyne.TextAlignCenter, fyne.TextStyle{}),
			container.NewHBox(layout.NewSpacer(), okButton, layout.NewSpacer()),
		)
		dialog.ShowCustomWithoutButtons("", content, window)
	})

	themeSwitch.OnTapped = func() {
		if isDarkTheme {
			myApp.Settings().SetTheme(&purpleTheme{base: theme.LightTheme()})
			themeSwitch.SetText("🌙")
			isDarkTheme = false
		} else {
			myApp.Settings().SetTheme(&purpleTheme{base: theme.DarkTheme()})
			themeSwitch.SetText("☀️")
			isDarkTheme = true
		}
	}

	buttonRow := container.NewHBox(
		layout.NewSpacer(),
		quoteBtn,
		clearBtn,
		layout.NewSpacer(),
	)

	topRow := container.NewHBox(
		infoBtn,
		layout.NewSpacer(),
		themeSwitch,
	)

	mainContent := container.NewBorder(
		container.NewVBox(
			widget.NewSeparator(),
			topRow,
			buttonRow,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewScroll(textEdit),
	)

	window.SetContent(mainContent)
	window.ShowAndRun()
}
