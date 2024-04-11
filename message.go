package messages

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrReservedKeyword      = errors.New("reserved keyword")
	ErrIdentifierInvalid    = errors.New("identifier must start with an uppercase letter and contain only letters and numbers")
	ErrUnsupportedFormat    = errors.New("format only supports 's' and 'd'")
	ErrVariableTypeMix      = errors.New("a variable can only be of one type")
	ErrDuplicateTranslation = errors.New("duplicate translation")
	ErrDuplicateIdentifier  = errors.New("duplicate identifier")

	identifierRe = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)

	varRe = regexp.MustCompile(`(?m)%\(([a-zA-Z]+)\)([a-z])`)
)

// NewLocalizedMessage creates a new message container.
func NewLocalizedMessage(identifier string, defaultMessage *Message) (*LocalizedMessage, error) {
	if !identifierRe.MatchString(identifier) {
		return nil, ErrIdentifierInvalid
	}

	return &LocalizedMessage{
		Identifier:   identifier,
		Default:      defaultMessage,
		Translations: make([]*Translation, 0),
	}, nil
}

func ParseMessage(raw string) (*Message, error) {
	msg := &Message{
		Message: raw,
		Vars:    make([]*Var, 0),
	}

	for _, varMatch := range varRe.FindAllStringSubmatch(raw, -1) {
		if isReservedKeyword(varMatch[1]) {
			return nil, ErrReservedKeyword
		}

		if varMatch[2] != "s" && varMatch[2] != "d" {
			return nil, fmt.Errorf("%q's var %q contains format %q: %w", raw, varMatch[1], varMatch[2], ErrUnsupportedFormat)
		}

		msgVar := &Var{
			Name: varMatch[1],
		}

		switch varMatch[2] {
		case "d":
			msgVar.Type = VarTypeInt
		case "s":
			msgVar.Type = VarTypeString
		}

		// Check if the var already exists with a different format.
		existing := msg.Var(varMatch[1])
		if existing != nil && existing.Type != msgVar.Type {
			return nil, fmt.Errorf("%q's var %q has type %q and %q: %w", raw, varMatch[1], existing.Type, msgVar.Type, ErrVariableTypeMix)
		}

		msg.Vars = append(msg.Vars, msgVar)
		msg.Message = strings.ReplaceAll(msg.Message, varMatch[0], "%"+varMatch[2])
	}

	return msg, nil
}

type Messages struct {
	// Name contains the capitalized filename without the extension.
	Name     string
	Messages []*LocalizedMessage
}

func (c *Messages) Add(m *LocalizedMessage) error {
	for _, msg := range c.Messages {
		if msg.Identifier == m.Identifier {
			return ErrDuplicateIdentifier
		}
	}

	c.Messages = append(c.Messages, m)

	return nil
}

func (c Messages) HasType(tp VarType) bool {
	for _, message := range c.Messages {
		if message.Default.HasType(tp) {
			return true
		}

		for _, tr := range message.Translations {
			if tr.Message.HasType(tp) {
				return true
			}
		}
	}

	return false
}

func (c Messages) HasTranslations() bool {
	for _, message := range c.Messages {
		if len(message.Translations) > 0 {
			return true
		}
	}

	return false
}

// LocalizedMessage contains a default message and optional translations by it's identifier.
type LocalizedMessage struct {
	Identifier string
	Default    *Message
	// Translations contains translations by locale.
	Translations []*Translation
}

type Translation struct {
	Locale  string
	Message *Message
}

func (l *LocalizedMessage) HasType(t VarType) bool {
	if l.Default.HasType(t) {
		return true
	}

	for _, tr := range l.Translations {
		if tr.Message.HasType(t) {
			return true
		}
	}

	return false
}

func (l *LocalizedMessage) AddTranslation(locale string, message *Message) error {
	if err := varTypesConsistent(l.Default, message); err != nil {
		return err
	}

	for _, tr := range l.Translations {
		if err := varTypesConsistent(tr.Message, message); err != nil {
			return err
		}

		if tr.Locale == locale {
			return fmt.Errorf("%w: locale = %q", ErrDuplicateTranslation, locale)
		}
	}

	l.Translations = append(l.Translations, &Translation{
		Locale:  locale,
		Message: message,
	})

	return nil
}

func (l *LocalizedMessage) UniqueVars() []*Var {
	vars := l.Default.UniqueVars()

	for _, tr := range l.Translations {
		uniqueVars := tr.Message.UniqueVars()

		for _, v := range uniqueVars {
			found := false
			for _, u := range vars {
				if u.Name == v.Name {
					found = true
					break
				}
			}

			if !found {
				vars = append(vars, v)
			}
		}
	}

	return vars
}

// Message is a single message and it's vars.
type Message struct {
	Message string
	Vars    []*Var
}

func (m *Message) UniqueVars() []*Var {
	vars := make([]*Var, 0)

	for _, v := range m.Vars {
		found := false
		for _, u := range vars {
			if u.Name == v.Name {
				found = true
				break
			}
		}
		if !found {
			vars = append(vars, v)
		}
	}

	return vars
}

func (m *Message) HasType(t VarType) bool {
	for _, v := range m.Vars {
		if v.Type == t {
			return true
		}
	}

	return false
}

func (m *Message) Var(name string) *Var {
	for _, v := range m.Vars {
		if v.Name == name {
			return v
		}
	}
	return nil
}

// Var is a variable in a translation.
type Var struct {
	Name string
	Type VarType
}

type VarType string

var (
	VarTypeString VarType = "string"
	VarTypeInt    VarType = "int"
)

// reservedKeywords contains all the reserved keywords. Variables and functions cannot have these names.
var reservedKeywords = []string{
	"ctx",
	"break", "default", "func", "interface", "select",
	"case", "defer", "go", "map", "struct",
	"chan", "else", "goto", "package", "switch",
	"const", "fallthrough", "if", "range", "type",
	"continue", "for", "import", "return", "var",
}

// isReservedKeyword checks if the given word is a reserved keyword in Go.
func isReservedKeyword(word string) bool {
	for _, keyword := range reservedKeywords {
		if strings.EqualFold(word, keyword) {
			return true
		}
	}
	return false
}

// varTypesConsistent checks if the variable types are consistent between the compared and target message.
func varTypesConsistent(comp *Message, target *Message) error {
	for _, compVar := range comp.Vars {
		for _, targetVar := range target.Vars {
			if compVar.Name == targetVar.Name && compVar.Type != targetVar.Type {
				return fmt.Errorf("variable %q has type %q and %q: %w", compVar.Name, compVar.Type, targetVar.Type, ErrVariableTypeMix)
			}
		}
	}

	return nil
}
