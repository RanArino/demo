package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("clerk_user_id").
			Unique(),
		field.String("email").
			Unique(),
		field.String("full_name"),
		field.String("username").
			Optional(),
		field.String("role").
			Default("user"),
		field.Int64("storage_used_bytes").
			Default(0),
		field.Int64("storage_quota_bytes").
			Default(5368709120), // 5GB
		field.String("status").
			Default("active"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("preferences", UserPreferences.Type).
			Unique(),
	}
}
