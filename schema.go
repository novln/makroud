package sqlxx

import (
	"fmt"
	"reflect"

	"github.com/oleiade/reflections"
)

// Schema is a model schema.
type Schema struct {
	Columns      map[string]Column
	Associations map[string]RelatedField
}

// Column is a database column
type Column struct {
	TableName    string
	Name         string
	PrefixedName string
}

// RelatedField represents an related field between two models.
type RelatedField struct {
	FK          Column
	FKReference Column
}

// GetSchema returns model's table columns, extracted by reflection.
// The returned map is modelFieldName -> table_name.column_name
func GetSchema(model Model) (*Schema, error) {
	fields, err := reflections.Fields(model)
	if err != nil {
		return nil, err
	}

	schema := &Schema{
		Columns:      map[string]Column{},
		Associations: map[string]RelatedField{},
	}

	for _, field := range fields {
		kind, err := reflections.GetFieldKind(model, field)
		if err != nil {
			return nil, err
		}

		// Associations

		if kind == reflect.Struct || kind == reflect.Ptr {
			relatedField, err := newRelatedField(model, field)
			if err != nil {
				return nil, err
			}

			schema.Associations[field] = relatedField

			continue
		}

		// Columns

		tag, err := reflections.GetFieldTag(model, field, SQLXStructTagName)
		if err != nil {
			return nil, err
		}

		col, err := newColumn(model, field, tag, false, false)
		if err != nil {
			return nil, err
		}

		schema.Columns[field] = col
	}

	return schema, nil
}

// newRelatedField creates a new related field.
func newRelatedField(model Model, field string) (RelatedField, error) {
	relatedField := RelatedField{}

	relatedValue, err := reflections.GetField(model, field)
	if err != nil {
		return relatedField, err
	}

	dbTag, err := reflections.GetFieldTag(model, field, SQLXStructTagName)
	if err != nil {
		return relatedField, err
	}

	tag, err := reflections.GetFieldTag(model, field, StructTagName)
	if err != nil {
		return relatedField, err
	}

	related := relatedValue.(Model)

	relatedField.FK, err = newColumn(model, field, dbTag, true, false)
	if err != nil {
		return relatedField, err
	}

	relatedField.FKReference, err = newColumn(related, field, tag, true, true)
	if err != nil {
		return relatedField, err
	}

	return relatedField, nil
}

// newColumn returns full column name from model, field and tag.
func newColumn(model Model, field string, tag string, isRelated bool, isReference bool) (Column, error) {
	// Retrieve the model type
	reflectType := reflect.ValueOf(model).Type()

	// If it's a pointer, we must get the elem to avoid double pointer errors
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}

	// Then we can safely cast
	reflected := reflect.New(reflectType).Interface().(Model)

	hasTag := len(tag) > 0

	// Build column name from tag or field
	column := tag
	if !hasTag {
		column = toSnakeCase(field)
	}

	// It's not a related field, early return
	if !isRelated {
		return Column{
			TableName:    reflected.TableName(),
			Name:         column,
			PrefixedName: fmt.Sprintf("%s.%s", reflected.TableName(), column),
		}, nil
	}

	// Reference primary key fields are "id" and "field_id"
	if isReference {
		column = "id"

		if hasTag {
			column = tag
		}

		return Column{
			TableName:    reflected.TableName(),
			Name:         column,
			PrefixedName: fmt.Sprintf("%s.%s", reflected.TableName(), column),
		}, nil
	}

	// It's a foreign key
	column = fmt.Sprintf("%s_id", column)
	if hasTag {
		column = tag
	}

	return Column{
		TableName:    reflected.TableName(),
		Name:         column,
		PrefixedName: fmt.Sprintf("%s.%s", reflected.TableName(), column),
	}, nil
}
