package terminal

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/x/ansi"
)

type richFormModel interface {
	tea.Model
	configure(width, height int, showHelp bool)
	handlesEscape() bool
}

func newRichForm(handler *InteractionHandler, request InteractionRequest, id uint64) (richFormModel, func() InteractionAnswer, error) {
	if request.Kind == InteractionSelect || request.Kind == InteractionMultiSelect {
		form := newRichListForm(request, id)
		return form, form.answer, nil
	}

	form, answer, err := handler.huhForm(request)
	if err != nil {
		return nil, nil, err
	}
	form.WithKeyMap(huh.NewDefaultKeyMap())
	form.SubmitCmd = func() tea.Msg { return richFormSubmittedMsg{id: id} }
	form.CancelCmd = func() tea.Msg { return richFormCancelledMsg{id: id} }
	return &richHuhForm{form: form}, answer, nil
}

type richHuhForm struct {
	form *huh.Form
}

func (form *richHuhForm) Init() tea.Cmd {
	return form.form.Init()
}

func (form *richHuhForm) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	updated, command := form.form.Update(message)
	form.form = updated.(*huh.Form)
	return form, command
}

func (form *richHuhForm) View() string {
	return form.form.View()
}

func (form *richHuhForm) configure(width, height int, showHelp bool) {
	form.form.WithWidth(width).WithHeight(height).WithShowHelp(showHelp)
}

func (form *richHuhForm) handlesEscape() bool {
	field, ok := form.form.GetFocusedField().(interface{ GetFiltering() bool })
	return ok && field.GetFiltering()
}

type richListForm struct {
	request InteractionRequest
	id      uint64
	multi   bool

	width     int
	height    int
	showHelp  bool
	cursor    int
	offset    int
	filtered  []int
	selected  map[int]bool
	filter    string
	filtering bool
	err       error
}

func newRichListForm(request InteractionRequest, id uint64) *richListForm {
	form := &richListForm{
		request:  request,
		id:       id,
		multi:    request.Kind == InteractionMultiSelect,
		filtered: make([]int, len(request.Options)),
		selected: make(map[int]bool),
	}
	for index := range request.Options {
		form.filtered[index] = index
	}
	if form.multi && request.HasDefault {
		cursorSet := false
		for index, option := range request.Options {
			for _, value := range request.Default.Values {
				if option.Value == value {
					form.selected[index] = true
					if !cursorSet {
						form.cursor = index
						cursorSet = true
					}
					break
				}
			}
		}
	} else if !form.multi && request.HasDefault {
		for index, option := range request.Options {
			if option.Value == request.Default.Value {
				form.cursor = index
				break
			}
		}
	}
	return form
}

func (*richListForm) Init() tea.Cmd {
	return nil
}

func (form *richListForm) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := message.(tea.WindowSizeMsg); ok {
		form.ensureVisible()
		return form, nil
	}
	key, ok := message.(tea.KeyMsg)
	if !ok {
		return form, nil
	}

	if form.filtering {
		switch key.String() {
		case "ctrl+c":
			return form, func() tea.Msg { return richFormCancelledMsg{id: form.id} }
		case "esc", "enter":
			form.filtering = false
		case "backspace":
			if len(form.filter) > 0 {
				runes := []rune(form.filter)
				form.filter = string(runes[:len(runes)-1])
				form.applyFilter()
			}
		default:
			if key.Type == tea.KeyRunes {
				form.filter += string(key.Runes)
				form.applyFilter()
			}
		}
		return form, nil
	}

	switch key.String() {
	case "ctrl+c":
		return form, func() tea.Msg { return richFormCancelledMsg{id: form.id} }
	case "esc":
		if form.filter != "" {
			form.filter = ""
			form.applyFilter()
		}
	case "/":
		form.filtering = true
	case "up", "k", "ctrl+k", "ctrl+p":
		form.move(-1)
	case "down", "j", "ctrl+j", "ctrl+n":
		form.move(1)
	case "home", "g":
		form.cursor = 0
		form.ensureVisible()
	case "end", "G":
		if len(form.filtered) > 0 {
			form.cursor = len(form.filtered) - 1
		}
		form.ensureVisible()
	case "ctrl+u":
		form.move(-max(form.optionRows()/2, 1))
	case "ctrl+d":
		form.move(max(form.optionRows()/2, 1))
	case "ctrl+a":
		if form.multi {
			form.toggleAll()
		}
	case " ", "x":
		if form.multi {
			form.toggleCurrent()
		}
	case "enter", "tab":
		if len(form.filtered) == 0 {
			return form, nil
		}
		if err := validateAnswer(form.request, form.answer()); err != nil {
			form.err = err
			form.ensureVisible()
			return form, nil
		}
		form.err = nil
		return form, func() tea.Msg { return richFormSubmittedMsg{id: form.id} }
	}
	return form, nil
}

func (form *richListForm) View() string {
	lines := make([]string, 0, form.height)
	lines = append(lines, splitRichListLine(form.request.Message, form.width)...)
	if form.request.Description != "" {
		lines = append(lines, splitRichListLine(form.request.Description, form.width)...)
	}
	if form.filtering || form.filter != "" {
		lines = append(lines, truncateRichListLine("/"+form.filter, form.width))
	}
	if form.err != nil {
		lines = append(lines, truncateRichListLine(form.err.Error(), form.width))
	}

	rows := form.optionRows()
	if len(form.filtered) == 0 {
		lines = append(lines, "No matches")
	} else {
		end := min(form.offset+rows, len(form.filtered))
		for position := form.offset; position < end; position++ {
			optionIndex := form.filtered[position]
			option := form.request.Options[optionIndex]
			label := option.Label
			if option.Description != "" {
				label += " - " + option.Description
			}
			prefix := "  "
			if form.multi {
				prefix = "[ ] "
				if form.selected[optionIndex] {
					prefix = "[x] "
				}
			}
			if position == form.cursor {
				prefix = "> " + prefix
			} else {
				prefix = "  " + prefix
			}
			lines = append(lines, prefix+truncateRichListLine(label, max(form.width-len(prefix), 1)))
		}
	}
	if form.showHelp {
		command := "enter select"
		if form.multi {
			command = "space toggle | ctrl+a all | enter confirm"
		}
		lines = append(lines, truncateRichListLine("up/down move | / filter | "+command+" | esc cancel", form.width))
	}
	return strings.Join(lines, "\n")
}

func (form *richListForm) configure(width, height int, showHelp bool) {
	form.width = max(width, 1)
	form.height = max(height, 1)
	form.showHelp = showHelp
	form.ensureVisible()
}

func (form *richListForm) handlesEscape() bool {
	return form.filtering || form.filter != ""
}

func (form *richListForm) answer() InteractionAnswer {
	if form.multi {
		values := make([]string, 0, len(form.selected))
		for index, option := range form.request.Options {
			if form.selected[index] {
				values = append(values, option.Value)
			}
		}
		return InteractionAnswer{Values: values}
	}
	if len(form.filtered) == 0 {
		return InteractionAnswer{}
	}
	return InteractionAnswer{Value: form.request.Options[form.filtered[form.cursor]].Value}
}

func (form *richListForm) applyFilter() {
	needle := strings.ToLower(strings.TrimSpace(form.filter))
	form.filtered = form.filtered[:0]
	for index, option := range form.request.Options {
		haystack := strings.ToLower(option.Label + " " + option.Value + " " + option.Description)
		if needle == "" || strings.Contains(haystack, needle) {
			form.filtered = append(form.filtered, index)
		}
	}
	form.cursor = 0
	form.offset = 0
	form.ensureVisible()
}

func (form *richListForm) move(delta int) {
	if len(form.filtered) == 0 {
		return
	}
	form.cursor = min(max(form.cursor+delta, 0), len(form.filtered)-1)
	form.ensureVisible()
}

func (form *richListForm) toggleCurrent() {
	if len(form.filtered) == 0 {
		return
	}
	index := form.filtered[form.cursor]
	form.selected[index] = !form.selected[index]
	if !form.selected[index] {
		delete(form.selected, index)
	}
}

func (form *richListForm) toggleAll() {
	allSelected := len(form.filtered) > 0
	for _, index := range form.filtered {
		if !form.selected[index] {
			allSelected = false
			break
		}
	}
	for _, index := range form.filtered {
		if allSelected {
			delete(form.selected, index)
		} else {
			form.selected[index] = true
		}
	}
}

func (form *richListForm) optionRows() int {
	headerRows := len(splitRichListLine(form.request.Message, form.width))
	if form.request.Description != "" {
		headerRows += len(splitRichListLine(form.request.Description, form.width))
	}
	if form.filtering || form.filter != "" {
		headerRows++
	}
	if form.err != nil {
		headerRows++
	}
	footerRows := 0
	if form.showHelp {
		footerRows = 1
	}
	return max(form.height-headerRows-footerRows, 1)
}

func (form *richListForm) ensureVisible() {
	if len(form.filtered) == 0 {
		form.cursor = 0
		form.offset = 0
		return
	}
	form.cursor = min(max(form.cursor, 0), len(form.filtered)-1)
	rows := form.optionRows()
	if form.cursor < form.offset {
		form.offset = form.cursor
	}
	if form.cursor >= form.offset+rows {
		form.offset = form.cursor - rows + 1
	}
	form.offset = min(max(form.offset, 0), max(len(form.filtered)-rows, 0))
}

func splitRichListLine(value string, width int) []string {
	value = stripTerminalControl(value)
	if value == "" {
		return nil
	}
	return strings.Split(wrapText(value, max(width, 1)), "\n")
}

func truncateRichListLine(value string, width int) string {
	return ansi.Truncate(stripTerminalControl(value), max(width, 1), "...")
}
