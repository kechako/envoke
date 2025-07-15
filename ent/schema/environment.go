package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Environment holds the schema definition for the Environment entity.
type Environment struct {
	ent.Schema
}

// Fields of the Environment.
func (Environment) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique().
			NotEmpty(),
		field.String("description").
			Optional(),
	}
}

// Edges of the Environment.
func (Environment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("variables", Variable.Type).Annotations(
			entsql.Annotation{
				OnDelete: entsql.Cascade,
			},
		),
	}
}
