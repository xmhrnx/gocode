package settingsview

import (
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dropdown"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/providers"
	"gocode/internal/settings"
	"gocode/ui/theme"
)

// customProviderForm adds a provider at any address speaking one of the
// supported APIs: a company gateway, a proxy, a server on the network.
type customProviderForm struct {
	page               *providersPage
	open               bool
	name, baseURL, key *input.Input
	models             *input.Input
	style              *dropdown.Select
	show, add, cancel  *giode.ButtonElement
}

var customStyles = []struct {
	style providers.Style
	label string
}{
	{providers.StyleOpenAI, "OpenAI-compatible"},
	{providers.StyleAnthropic, "Anthropic Messages"},
	{providers.StyleGemini, "Google Gemini"},
}

func newCustomProviderForm(page *providersPage) *customProviderForm {
	f := &customProviderForm{page: page}
	f.name = giode.Input.Text("Name, e.g. Company gateway", fieldStyle(-1))
	f.baseURL = giode.Input.Text("Base URL, e.g. http://192.168.1.20:8000/v1", fieldStyle(-1))
	f.key = giode.Input.Text("API key (optional)", fieldStyle(-1)).Password(true)
	f.models = giode.Input.Text("Model IDs, comma separated (optional if the server lists them)", fieldStyle(-1))
	labels := make([]string, len(customStyles))
	for i, s := range customStyles {
		labels[i] = s.label
	}
	f.style = giode.Input.Select(fieldStyle(220), labels...)
	f.show = smallButton("Add custom provider", false).Icon(theme.Icon("plus", 14, "")).OnClick(func() { f.open = true })
	f.add = smallButton("Add provider", true).OnClick(f.submit)
	f.cancel = smallButton("Cancel", false).OnClick(func() { f.open = false })
	return f
}

func (f *customProviderForm) submit() {
	name, baseURL := strings.TrimSpace(f.name.Text()), strings.TrimSpace(f.baseURL.Text())
	if name == "" || baseURL == "" {
		f.page.services.Error("A custom provider needs a name and a base URL.")
		return
	}
	f.page.addCustom(settings.ProviderSettings{
		Name: name, BaseURL: baseURL, Style: string(customStyles[f.style.Selected()].style),
		ExtraModels: parseModelList(f.models.Text()),
	}, strings.TrimSpace(f.key.Text()))
	for _, field := range []*input.Input{f.name, f.baseURL, f.key, f.models} {
		field.SetText("")
	}
	f.open = false
}

func (f *customProviderForm) Layout(gtx giode.Context) giode.Dimensions {
	if !f.open {
		return f.show.Layout(gtx)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 12, Padding: giode.UniformInset(16), BorderRadius: 10,
		Background: theme.Raised, BorderWidth: 1, BorderColor: theme.SoftBorder},
		giode.Text("Custom provider", giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightSemiBold}),
		field("Name", f.name),
		field("API", f.style),
		field("Base URL", f.baseURL),
		field("API key", f.key),
		field("Models", f.models),
		giode.Row(giode.Styles{Gap: 8}, f.add, f.cancel),
	).Layout(gtx)
}
