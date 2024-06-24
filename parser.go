package staticmessages

import (
	"errors"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

var (
	ErrYamlNameInvalid       = errors.New("yml name invalid")
	ErrYamlDefinitionInvalid = errors.New("yml definition is invalid")
)

// Parse parses yml messages from r.
func Parse(name string, r io.Reader) (*Messages, error) {
	if len(name) == 0 {
		return nil, ErrYamlNameInvalid
	}

	// Capitalize the first letter of the name.
	ru, size := utf8.DecodeRuneInString(name)
	name = string(unicode.ToUpper(ru)) + name[size:]

	var node yaml.Node
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&node); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrYamlDefinitionInvalid, err)
	}

	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return nil, fmt.Errorf("%w: expected yaml.MappingNode got Document node without content", ErrYamlDefinitionInvalid)
		}

		if len(node.Content) > 1 {
			return nil, fmt.Errorf("%w: expected yaml.MappingNode got Document node with more then 1 child", ErrYamlDefinitionInvalid)
		}

		node = *node.Content[0]
	}

	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%w: expected yaml.MappingNode got: %d", ErrYamlDefinitionInvalid, node.Kind)
	}

	messages := &Messages{
		Name:     name,
		Messages: make([]*LocalizedMessage, 0),
	}

	// We parse the yaml manually into a yaml.Node to maintain the ordering of the fields as defined in r.
	// If we would use a map the ordering is not guaranteed.
	//
	// The following yaml:
	//
	// MessageOne:
	//   default: Hello 1
	// MessageTwo:
	//   default: Hello 2
	//   nl: Hallo 2
	//
	// Results into the following items in node.Content:
	//
	// Type					Index							Content
	// (yaml.ScalarNode) 	node.Content[0]					MessageOne
	// (yaml.MappingNode) 	node.Content[1]					default:Hello 1
	// (yaml.ScalarNode) 	node.Content[2]					MessageTwo
	// (yaml.MappingNode) 	node.Content[3]					default: Hello 2\nnl: Hallo 2
	//
	// This means we need to make hops in our for loop of 2, hence the i += 2.
	for i := 0; (i + 1) < len(node.Content); i += 2 {
		identifier := node.Content[i]
		spec := node.Content[i+1]

		if identifier.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("%w: expected yaml.ScalarNode for node %#v", ErrYamlDefinitionInvalid, identifier)
		}

		if spec.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("%w: expected yaml.MappingNode for node %#v", ErrYamlDefinitionInvalid, identifier)
		}

		loc, err := parseMessage(identifier.Value, spec)
		if err != nil {
			return nil, err
		}

		messages.Messages = append(messages.Messages, loc)
	}

	return messages, nil
}

func parseMessage(identifier string, spec *yaml.Node) (*LocalizedMessage, error) {
	var defaultMessage *Message
	translations := make([]*Translation, 0)
	var err error

	for i := 0; (i + 1) < len(spec.Content); i += 2 {
		key := spec.Content[i]
		value := spec.Content[i+1]

		if key.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("%w: expected yaml.ScalarNode for node %#v", ErrYamlDefinitionInvalid, key)
		}

		if value.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("%w: expected yaml.ScalarNode for node %#v", ErrYamlDefinitionInvalid, value)
		}

		if key.Value == "default" {
			defaultMessage, err = ParseMessage(value.Value)
			if err != nil {
				return nil, fmt.Errorf("%w: error parsing node: %#v: %v", ErrYamlDefinitionInvalid, value, err)
			}
		} else {
			translation, err := ParseMessage(value.Value)
			if err != nil {
				return nil, fmt.Errorf("%w: error parsing node: %#v: %v", ErrYamlDefinitionInvalid, value, err)
			}

			translations = append(translations, &Translation{
				Locale:  key.Value,
				Message: translation,
			})
		}
	}

	if defaultMessage == nil {
		return nil, fmt.Errorf("%w: expected default message", ErrYamlDefinitionInvalid)
	}

	loc, err := NewLocalizedMessage(identifier, defaultMessage)
	if err != nil {
		return nil, fmt.Errorf("%w: error creating localized message: %v", ErrYamlDefinitionInvalid, err)
	}

	for _, translation := range translations {
		err = loc.AddTranslation(translation.Locale, translation.Message)
		if err != nil {
			return nil, fmt.Errorf("%w: error adding translation: %v", ErrYamlDefinitionInvalid, err)
		}
	}

	return loc, nil
}
