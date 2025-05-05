package tui

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/rivo/tview"
)

func highlightCode(code, language string) string {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// Explicitly switch to the robust terminal256 formatter for broader compatibility.
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var buff bytes.Buffer
	err = formatter.Format(&buff, style, iterator)
	if err != nil {
		return code
	}

	return tview.TranslateANSI(buff.String())
}

func DetectLanguage(scriptContent string) string {
    lexer := lexers.Analyse(scriptContent)
    if lexer == nil {
        lexer = lexers.Match(scriptContent)
    }
    if lexer == nil {
        return "bash" // clearly fallback to bash
    }
    return lexer.Config().Name
}