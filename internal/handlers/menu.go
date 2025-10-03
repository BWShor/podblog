package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/labstack/echo/v4"
)

type MenuItem struct {
	ID    string
	Title string
}

func MenuHandler(c echo.Context) error {
	contentDir := "web/content"
	items := []MenuItem{}

	entries, err := os.ReadDir(contentDir)
	if err != nil {
		fmt.Println("Error reading content dir:", err)
		return err
	}

	titleCase := cases.Title(language.English)

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Println(" -", entry.Name())
			rawName := entry.Name()
			// Convert dashes/underscores to spaces for display
			displayName := strings.ReplaceAll(rawName, "-", " ")
			title := titleCase.String(displayName)

			// Normalize ID for URL (lowercase, replace spaces with dashes)
			normalizedID := strings.ToLower(strings.ReplaceAll(displayName, " ", "-"))
			fmt.Println("   Normalized ID:", normalizedID, "Title:", title)

			items = append(items, MenuItem{
				ID:    normalizedID,
				Title: title,
			})
		}
	}

	// Inline HTMX menu template
	tmpl := `
	<nav>
		<ul>
			{{range .Items}}
			<li>
				<a hx-get="/page/{{.ID}}/content" 
				hx-target="#content" 
				hx-swap="innerHTML" 
				hx-push-url="/page/{{.ID}}">
					{{.Title}}
				</a>
			</li>
			{{end}}
		</ul>
	</nav>`

	t, err := template.New("menu").Parse(tmpl)
	if err != nil {
		return err
	}

	nonce, _ := c.Get("cspNonce").(string)

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return t.Execute(c.Response(), struct {
		Items []MenuItem
		Nonce string
	}{
		Items: items,
		Nonce: nonce,
	})
}
