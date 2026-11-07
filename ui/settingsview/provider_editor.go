package settingsview

import (
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/providers"
	"gocode/ui/theme"

	"gocode/ui/components"
)

// providerEditor edits one provider: its key, address and extra models,
// and tests the connection by listing models.
type providerEditor struct {
	page    *providersPage
	id      string
	key     *input.Input
	baseURL *input.Input
	models  *input.Input
	saveKey *giode.ButtonElement
	test    *giode.ButtonElement
	getKey  *giode.ButtonElement
	remove  *giode.ButtonElement
}

func newProviderEditor(page *providersPage, def providers.Definition) *providerEditor {
	s := page.services
	e := &providerEditor{page: page, id: def.ID}
	e.key = giode.Input.Text(keyPlaceholder(def), fieldStyle(-1)).Password(true).Submit(func(string) { e.storeKey() })
	e.baseURL = giode.Input.Text(def.BaseURL, fieldStyle(-1))
	e.baseURL.SetText(def.BaseURL)
	e.baseURL.OnChange(func(text string) {
		providerSettings(s, def.ID).BaseURL = strings.TrimSpace(text)
		s.SettingsChanged()
	})
	e.models = giode.Input.Text("Extra model IDs, comma separated", fieldStyle(-1))
	if ps, ok := s.Settings.Provider(def.ID); ok {
		e.models.SetText(strings.Join(ps.ExtraModels, ", "))
	}
	e.models.OnChange(func(text string) {
		providerSettings(s, def.ID).ExtraModels = parseModelList(text)
		s.SettingsChanged()
	})
	components.AddEditMenu(s.App, e.key, components.EditOptions{Secret: true})
	components.AddEditMenu(s.App, e.baseURL, components.EditOptions{})
	components.AddEditMenu(s.App, e.models, components.EditOptions{})
	e.saveKey = smallButton("Save key", true).OnClick(e.storeKey)
	e.test = smallButton("Test connection", false).OnClick(func() { page.test(def.ID) })
	e.getKey = smallButton("Get a key", false).OnClick(func() { page.openURL(def.KeyURL) })
	e.remove = smallButton("Remove provider", false).OnClick(func() { page.removeCustom(def.ID) })
	return e
}

func (e *providerEditor) storeKey() {
	value := strings.TrimSpace(e.key.Text())
	if value == "" {
		return
	}
	e.key.SetText("")
	e.page.saveKey(e.id, value)
}

func (e *providerEditor) Layout(gtx giode.Context) giode.Dimensions {
	def, _ := e.page.find(e.id)
	keyActions := []giode.Element{e.saveKey}
	if def.KeyURL != "" {
		keyActions = append(keyActions, e.getKey)
	}
	actions := []giode.Element{e.test}
	if def.Category == "custom" {
		actions = append(actions, e.remove)
	}
	rows := []giode.Element{
		field("API key", giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, e.key), giode.Row(giode.Styles{Gap: 6}, keyActions...))),
		field("Base URL", e.baseURL),
		field("Extra models", e.models),
	}
	if def.Notes != "" {
		rows = append(rows, giode.Text(def.Notes, theme.Caption()))
	}
	rows = append(rows, giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter}, append(actions, e.page.testStatus(e.id))...))
	return giode.Stack(giode.Styles{Width: -1, Gap: 12, Padding: giode.Inset{Left: 16, Right: 16, Bottom: 16, Top: 4}}, rows...).Layout(gtx)
}

func keyPlaceholder(def providers.Definition) string {
	if def.KeyOptional {
		return "Optional for this server"
	}
	return "Paste your " + def.Name + " API key"
}

func field(label string, control giode.Element) giode.Element {
	return giode.Stack(giode.Styles{Width: -1, Gap: 5},
		giode.Text(label, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny, FontWeight: giode.FontWeightSemiBold}),
		control)
}

func smallButton(label string, primary bool) *giode.ButtonElement {
	if primary {
		return giode.Button(label, giode.Styles{Background: theme.Primary, Color: theme.OnPrimary, BorderRadius: 7,
			FontSize: theme.Sizes.Small, Padding: giode.SymmetricInset(12, 6)})
	}
	return giode.Button(label, giode.Styles{BorderColor: theme.Border, Color: theme.Text, BorderRadius: 7,
		FontSize: theme.Sizes.Small, Padding: giode.SymmetricInset(12, 6)}).Variant(giode.VariantOutlined)
}
