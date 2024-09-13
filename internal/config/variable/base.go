package variable

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"slices"
)

type Type string

const (
	Scalar Type = "scalar"
	Map    Type = "map"
	Slice  Type = "slice"
)

type VariablesMap map[string]Variable

// MergeVariablesMaps merges multiple VariablesMap into a single VariablesMap. The last map will override the previous
// ones. Only the keys on the first level are merged.
func MergeVariablesMaps(maps ...VariablesMap) VariablesMap {
	var result = make(VariablesMap)

	for _, m := range maps {
		for key, val := range m {
			result[key] = val
		}
	}

	return result
}

type TransformValueFunc func(value any) (any, error)

type Variable interface {
	Type() Type
	TransformValue(f TransformValueFunc) (any, error)
	ReferencedComponents() []string
}

type baseVariable struct {
	typ Type
}

func (v *baseVariable) Type() Type {
	return v.typ
}

func (vl *VariablesMap) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping node")
	}

	*vl = make(map[string]Variable, len(value.Content)/2)

	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i]
		val := value.Content[i+1]

		pVal, err := parseField(val)
		if err != nil {
			return err
		}
		(*vl)[key.Value] = pVal
	}

	return nil
}

func parseField(val *yaml.Node) (Variable, error) {
	switch val.Kind {
	case yaml.ScalarNode:
		return NewScalarVariable(val.Value)
	case yaml.MappingNode:
		var elements = make(map[string]Variable, len(val.Content)/2)
		for i := 0; i < len(val.Content); i += 2 {
			key := val.Content[i]
			val := val.Content[i+1]

			pVal, err := parseField(val)
			if err != nil {
				return nil, err
			}
			elements[key.Value] = pVal
		}

		return NewMapVariable(elements), nil
	case yaml.SequenceNode:
		var elements = make([]Variable, 0, len(val.Content))

		for _, v := range val.Content {
			pVal, err := parseField(v)
			if err != nil {
				return nil, err
			}
			elements = append(elements, pVal)
		}
		return NewSliceVariable(elements), nil
	default:
		return nil, fmt.Errorf("unsupported variable type: %s", val.ShortTag())
	}
}

func (vl *VariablesMap) Transform(f TransformValueFunc) (map[string]any, error) {
	var data = make(map[string]any, len(*vl))

	for key, element := range *vl {
		dat, err := element.TransformValue(f)
		if err != nil {
			return nil, err
		}
		data[key] = dat
	}

	return data, nil
}

func (vl *VariablesMap) ListReferencedComponents() []string {
	var references []string

	for _, v := range *vl {
		references = append(references, v.ReferencedComponents()...)
	}

	return slices.Compact(references)
}
