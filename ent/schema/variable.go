package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Variable holds the schema definition for the Variable entity.
type Variable struct {
	ent.Schema
}

// Fields of the Variable.
func (Variable) Fields() []ent.Field {
	return []ent.Field{
		field.Int("environment_id"),
		field.String("name").
			NotEmpty(),
		field.String("value"),
		field.String("comment").
			Optional(),
		field.Bool("expand").
			Optional(),
	}
}

// Edges of the Variable.
func (Variable) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("environment", Environment.Type).
			Ref("variables").
			Unique().
			Required().
			Field("environment_id"),
	}
}

func (Variable) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("environment_id", "name").
			Unique(),
	}
}
