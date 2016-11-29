package sqlxx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/serenize/snaker"
)

// ----------------------------------------------------------------------------
// Tag
// ----------------------------------------------------------------------------

// Tag is a field tag.
type Tag map[string]string

// Get returns value for the given key or zero value.
func (t Tag) Get(key string) string {
	v, _ := t[key]
	return v
}

// ----------------------------------------------------------------------------
// Tags
// ----------------------------------------------------------------------------

// Tags are field tags.
type Tags map[string]Tag

// Get returns the given tag.
func (t Tags) Get(name string) (Tag, error) {
	tag, ok := t[name]
	if !ok {
		return nil, fmt.Errorf("tag %s does not exist", name)
	}

	return tag, nil
}

// GetByKey is a convenient shortcuts to get the value for a given tag key.
func (t Tags) GetByKey(name string, key string) string {
	if tag, err := t.Get(name); err == nil {
		if v := tag.Get(key); len(v) != 0 {
			return v
		}
	}

	return ""
}

// makeTags returns field tags formatted.
func makeTags(structField reflect.StructField) Tags {
	tags := Tags{}

	rawTags := getFieldTags(structField, SupportedTags...)

	for k, v := range rawTags {
		splits := strings.Split(v, ";")

		tags[k] = map[string]string{}

		// Properties
		vals := []string{}
		for _, s := range splits {
			if len(s) != 0 {
				vals = append(vals, strings.TrimSpace(s))
			}
		}

		// Key / value
		for _, v := range vals {
			splits = strings.Split(v, ":")
			length := len(splits)

			if length == 0 {
				continue
			}

			// format: db:"field_name" -> "field" -> "field_name"
			if k == SQLXStructTagName {
				tags[k]["field"] = strings.TrimSpace(splits[0])
				continue
			}

			// Typically, we have single property like "default", "ignored", etc.
			// To be consistent, we add true/false string values.
			if length == 1 {
				tags[k][strings.TrimSpace(splits[0])] = "true"
				continue
			}

			// Typical key / value
			if length == 2 {
				tags[k][strings.TrimSpace(splits[0])] = strings.TrimSpace(splits[1])
			}
		}
	}

	return tags
}

func getFieldTags(structField reflect.StructField, names ...string) map[string]string {
	tags := map[string]string{}

	for _, name := range names {
		if _, ok := tags[name]; !ok {
			tags[name] = structField.Tag.Get(name)
		}
	}

	return tags
}

// ----------------------------------------------------------------------------
// Meta
// ----------------------------------------------------------------------------

// Meta are low level field metadata.
type Meta struct {
	Name  string
	Field reflect.StructField
	Type  reflect.Type
	Tags  Tags
}

func makeMeta(field reflect.StructField) Meta {
	var (
		fieldName = field.Name
		fieldType = field.Type
	)

	if field.Type.Kind() == reflect.Ptr {
		fieldType = field.Type.Elem()
	}

	return Meta{
		Name:  fieldName,
		Field: field,
		Type:  fieldType,
		Tags:  makeTags(field),
	}
}

// ----------------------------------------------------------------------------
// Field
// ----------------------------------------------------------------------------

// Field is a field.
type Field struct {
	// Struct field name.
	Name string

	// Struct field metadata (reflect data).
	Meta Meta

	// Struct field tags.
	Tags Tags

	// TableName is the database table name.
	TableName string

	// ColumnName is the database column name.
	ColumnName string

	// Is a primary key?
	IsPrimary bool
}

// ColumnPath returns the column name prefixed with the table name.
func (f Field) ColumnPath() string {
	return fmt.Sprintf("%s.%s", f.TableName, f.ColumnName)
}

// newField returns full column name from model, field and tag.
func newField(model Model, meta Meta) (Field, error) {
	tags := makeTags(meta.Field)

	var columnName string

	if dbName := tags.GetByKey(SQLXStructTagName, "field"); len(dbName) != 0 {
		columnName = dbName
	} else {
		columnName = snaker.CamelToSnake(meta.Name)
	}

	return Field{
		Name:       meta.Name,
		Meta:       meta,
		Tags:       tags,
		TableName:  model.TableName(),
		ColumnName: columnName,
	}, nil
}

// newForeignKeyField returns foreign key field.
func newForeignKeyField(model Model, meta Meta) (Field, error) {
	field, err := newField(model, meta)
	if err != nil {
		return Field{}, err
	}

	// Defaults to "fieldname_id"
	field.ColumnName = fmt.Sprintf("%s_id", field.ColumnName)

	// Get the SQLX one if any.
	if customName := field.Tags.GetByKey(SQLXStructTagName, "field"); len(customName) != 0 {
		field.ColumnName = customName
	}

	return field, nil
}

// newForeignKeyReferenceField returns a foreign key reference field.
func newForeignKeyReferenceField(referencedModel Model, name string) (Field, error) {
	reflectType := reflectType(referencedModel)

	reflected := reflect.New(reflectType).Interface().(Model)

	f, ok := reflectType.FieldByName(name)
	if !ok {
		return Field{}, fmt.Errorf("Field %s does not exist", name)
	}

	meta := Meta{Name: name, Field: f}

	field, err := newField(reflected, meta)
	if err != nil {
		return Field{}, err
	}

	return field, nil
}

// ----------------------------------------------------------------------------
// Relation
// ----------------------------------------------------------------------------

// Relation represents an related field between two models.
type Relation struct {
	Model     Model
	Schema    Schema
	Type      RelationType
	FK        Field
	Reference Field
}

// newRelatedField creates a new related field.
func newRelation(model Model, meta Meta, typ RelationType) (Relation, error) {
	var err error

	relation := Relation{Type: typ}

	relation.FK, err = newForeignKeyField(model, meta)
	if err != nil {
		return relation, err
	}

	relation.Model = getModelType(meta.Type)

	schema, err := GetSchema(relation.Model)
	if err != nil {
		return relation, err
	}

	relation.Schema = schema

	relation.Reference, err = newForeignKeyReferenceField(relation.Model, "ID")
	if err != nil {
		return relation, err
	}

	return relation, nil
}

// getRelationType returns RelationType for the given reflect.Type.
func getRelationType(typ reflect.Type) RelationType {
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()

		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if _, isModel := reflect.New(typ).Interface().(Model); isModel {
			return RelationTypeManyToOne
		}

		return RelationTypeUnknown
	}

	if _, isModel := reflect.New(typ).Interface().(Model); isModel {
		return RelationTypeOneToMany
	}

	return RelationTypeUnknown
}
