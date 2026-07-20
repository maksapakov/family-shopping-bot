package matrix

import (
	"fmt"
	"strings"
	"testing"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
)

func TestFormatList(t *testing.T) {
	list := domain.NewShoppingList("list-1", "chat-1")
	list.AddItem(domain.NewItem("item-1", "Milk", "mom", domain.LocationProducts))
	rendered := list.Render()
	rendered.Items[0].ClickURL = "http://localhost:8181/toggle"
	rendered.UndoURL = "http://localhost:8181/undo"

	plain, html := formatList(rendered)
	wantPlain := fmt.Sprintf("☐ Milk\nundo\n")
	if plain != wantPlain {
		t.Fatalf("plain %s isn't match with expected %s", plain, wantPlain)
	}
	if !strings.Contains(html, "Milk") {
		t.Fatalf("html %s isn't contain %s", html, "Milk")
	}
	if !strings.Contains(html, "http://localhost:8181/toggle") {
		t.Fatalf("html %s isn't contain %s", html, "http://localhost:8181/toggle")
	}
	if !strings.Contains(html, "http://localhost:8181/undo") {
		t.Fatalf("html %s isn't contain %s", html, "http://localhost:8181/undo")
	}
}
