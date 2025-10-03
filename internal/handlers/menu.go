// internal/handlers/menu.go
package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

/* --------------------------------------------------------------------------- */
/* 1. Data structures -------------------------------------------------------- */
/* --------------------------------------------------------------------------- */

type MenuNode struct {
	ID       string     // slug (empty if no index.html)
	Title    string     // display title
	Children []MenuNode // nested items
}

// MenuOrder maps a node ID to the ordered list of its children.
// The special key "root" represents the top‑level menu.
type MenuOrder map[string][]string

/* --------------------------------------------------------------------------- */
/* 2. Helper: load / create ordering file ----------------------------------- */
/* --------------------------------------------------------------------------- */

const menuIndexPath = "menuindex.yml"

// LoadMenuOrder reads the YAML file and returns a MenuOrder.
// If the file cannot be read or parsed, it returns an empty map and the error.
func LoadMenuOrder(path string) (MenuOrder, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var order MenuOrder
	if err := yaml.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return order, nil
}

// createDefaultMenuIndex writes a menuindex.yml that reflects the current
// directory tree.  It only runs when the file does NOT exist.
// The file is written in a human‑readable, sorted order.
func createDefaultMenuIndex(path string, nodes []MenuNode) error {
	// If the file already exists, do nothing.
	if _, err := os.Stat(path); err == nil {
		return nil // nothing to do
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat menu file: %w", err)
	}

	order := make(MenuOrder)
	order["root"] = make([]string, 0, len(nodes))

	var walk func(node MenuNode)
	walk = func(node MenuNode) {
		if node.ID != "" { // only pages get an ID
			order["root"] = append(order["root"], node.ID)
		}
		if len(node.Children) > 0 {
			childIDs := make([]string, 0, len(node.Children))
			for _, c := range node.Children {
				if c.ID != "" {
					childIDs = append(childIDs, c.ID)
				}
			}
			order[node.ID] = childIDs
		}
		for _, c := range node.Children {
			walk(c)
		}
	}
	for _, n := range nodes {
		walk(n)
	}

	data, err := yaml.Marshal(order)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

/* --------------------------------------------------------------------------- */
/* 3. Recursive tree builder ----------------------------------------------- */
/* --------------------------------------------------------------------------- */

func buildMenuTree(dir string, rel string, order MenuOrder) (MenuNode, error) {
	titleCase := cases.Title(language.English)

	node := MenuNode{
		Title: titleCase.String(strings.ReplaceAll(filepath.Base(dir), "-", " ")),
	}

	// If this dir has an index.html, treat it as a page with an ID
	indexPath := filepath.Join(dir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		slug := strings.ToLower(strings.ReplaceAll(rel, string(os.PathSeparator), "-"))
		node.ID = slug
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return node, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			childRel := filepath.Join(rel, entry.Name())
			childDir := filepath.Join(dir, entry.Name())
			childNode, err := buildMenuTree(childDir, childRel, order)
			if err != nil {
				return node, err
			}
			node.Children = append(node.Children, childNode)
		}
	}

	// Re‑order children if an entry exists in the order map
	if order != nil {
		if childIDs, ok := order[slug(rel)]; ok {
			// Build a map for quick lookup
			childMap := make(map[string]MenuNode, len(node.Children))
			for _, c := range node.Children {
				childMap[c.ID] = c
			}
			ordered := make([]MenuNode, 0, len(node.Children))
			for _, childID := range childIDs {
				if child, ok := childMap[childID]; ok {
					ordered = append(ordered, child)
					delete(childMap, childID)
				}
			}
			// Append any remaining children that weren't listed
			for _, remaining := range childMap {
				ordered = append(ordered, remaining)
			}
			node.Children = ordered
		}
	}

	return node, nil
}

func buildMenu(contentDir string, order MenuOrder) ([]MenuNode, error) {
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		return nil, err
	}

	var nodes []MenuNode
	for _, entry := range entries {
		if entry.IsDir() {
			rel := entry.Name()
			dir := filepath.Join(contentDir, entry.Name())
			node, err := buildMenuTree(dir, rel, order)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
	}

	// Re‑order the top‑level nodes
	if order != nil {
		if childIDs, ok := order["root"]; ok {
			childMap := make(map[string]MenuNode, len(nodes))
			for _, c := range nodes {
				childMap[c.ID] = c
			}
			ordered := make([]MenuNode, 0, len(nodes))
			for _, childID := range childIDs {
				if child, ok := childMap[childID]; ok {
					ordered = append(ordered, child)
					delete(childMap, childID)
				}
			}
			for _, remaining := range childMap {
				ordered = append(ordered, remaining)
			}
			nodes = ordered
		}
	}

	return nodes, nil
}

/* --------------------------------------------------------------------------- */
/* 4. Template rendering ----------------------------------------------------- */
/* --------------------------------------------------------------------------- */

func MenuHandler(c echo.Context) error {
	// 1. Build the tree *without* an order first
	nodes, err := buildMenu("web/content", nil)
	if err != nil {
		return c.String(http.StatusInternalServerError,
			fmt.Sprintf("menu build error: %v", err))
	}

	// 2. Ensure a menuindex.yml exists (creates default if missing)
	if err := createDefaultMenuIndex(menuIndexPath, nodes); err != nil {
		// We don't want to fail the request if writing the file fails,
		// but we do want to log it.
		log.Printf("failed to create default menuindex.yml: %v", err)
	}

	// 3. Load the ordering file (now guaranteed to exist)
	order, err := LoadMenuOrder(menuIndexPath)
	if err != nil {
		// If it still fails (e.g. syntax error), fall back to empty order
		log.Printf("menu index load error: %v – using natural order", err)
		order = make(MenuOrder)
	}

	// 4. Re‑build the tree with the order applied
	nodes, err = buildMenu("web/content", order)
	if err != nil {
		return c.String(http.StatusInternalServerError,
			fmt.Sprintf("menu build error: %v", err))
	}

	// -----------------------------------------------------------------------
	// Render the menu template
	// -----------------------------------------------------------------------
	const tmpl = `
<nav>
  <ul>
    {{template "nodes" .Menu}}
  </ul>
</nav>

{{define "nodes"}}
  {{range .}}
    <li>
      {{if .ID}}
        <a hx-get="/page/{{.ID}}/content"
           hx-target="#content"
           hx-swap="innerHTML"
           hx-push-url="/page/{{.ID}}">
          {{.Title}}
        </a>
      {{else}}
        <span class="menu-heading">{{.Title}}</span>
      {{end}}
      {{if .Children}}
        <ul>
          {{template "nodes" .Children}}
        </ul>
      {{end}}
    </li>
  {{end}}
{{end}}
`

	t, err := template.New("menu").Parse(tmpl)
	if err != nil {
		return c.String(http.StatusInternalServerError,
			fmt.Sprintf("template parse error: %v", err))
	}

	nonce, _ := c.Get("cspNonce").(string)

	// Echo's Response writer is already an http.ResponseWriter,
	// so we can write directly to it.
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)

	return t.Execute(c.Response(), struct {
		Menu  []MenuNode
		Nonce string
	}{
		Menu:  nodes,
		Nonce: nonce,
	})
}

/* --------------------------------------------------------------------------- */
/* 5. Utility: slug conversion ---------------------------------------------- */
/* --------------------------------------------------------------------------- */

func slug(rel string) string {
	// Convert Windows path separators to URL separators
	rel = strings.ReplaceAll(rel, string(os.PathSeparator), "-")
	return strings.ToLower(rel)
}

/* --------------------------------------------------------------------------- */
/* 6. (Optional) Test helper – not used by the server ---------------------- */
/* --------------------------------------------------------------------------- */

// RenderMenuToBuffer is a small helper that can be used in tests or debugging.
// It returns the rendered menu as a string.
func RenderMenuToBuffer() (string, error) {
	nodes, err := buildMenu("web/content", nil)
	if err != nil {
		return "", err
	}
	if err := createDefaultMenuIndex(menuIndexPath, nodes); err != nil {
		return "", err
	}
	order, err := LoadMenuOrder(menuIndexPath)
	if err != nil {
		return "", err
	}
	nodes, err = buildMenu("web/content", order)
	if err != nil {
		return "", err
	}

	const tmpl = `
<nav>
  <ul>
    {{template "nodes" .Menu}}
  </ul>
</nav>

{{define "nodes"}}
  {{range .}}
    <li>
      {{if .ID}}
        <a href="/page/{{.ID}}/content">{{.Title}}</a>
      {{else}}
        <span>{{.Title}}</span>
      {{end}}
      {{if .Children}}
        <ul>
          {{template "nodes" .Children}}
        </ul>
      {{end}}
    </li>
  {{end}}
{{end}}
`
	t, err := template.New("menu").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, struct{ Menu []MenuNode }{Menu: nodes}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
