package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserPreferences holds the schema definition for the UserPreferences entity.
type UserPreferences struct {
	ent.Schema
}

// Fields of the UserPreferences.
func (UserPreferences) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("theme").
			Default("auto"),
		field.String("language").
			Default("en"),
		field.String("timezone").
			Optional(),
		field.JSON("canvas_settings", map[string]interface{}{}).
			Optional(),
		field.JSON("notification_settings", map[string]interface{}{}).
			Optional(),
		field.JSON("accessibility_settings", map[string]interface{}{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the UserPreferences.
func (UserPreferences) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("preferences").
			Unique().
			Required(),
	}
}
