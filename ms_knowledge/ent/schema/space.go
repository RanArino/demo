package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Space holds the schema definition for the Space entity.
type Space struct {
	ent.Schema
}

// Fields of the Space.
func (Space) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title"),
		field.Text("description").
			Optional(),
		field.String("icon").
			Optional(),
		field.String("cover_image").
			Optional(),
		field.Strings("keywords").
			Optional(),
		field.UUID("owner_id", uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.UUID("created_by", uuid.UUID{}),
		field.Time("last_updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.UUID("last_updated_by", uuid.UUID{}).
			Optional(),
		field.Int("document_count").
			Default(0),
		field.Int64("total_size_bytes").
			Default(0),
		field.Int64("storage_quota_bytes").
			Default(1073741824), // 1GB
		field.String("access_level").
			Default("private"),
		field.Bool("guest_access_enabled").
			Default(false),
		field.Time("guest_access_expiry").
			Optional().
			Nillable(),
		field.String("status").
			Default("active"),
		field.String("processing_status").
			Default("idle"),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Space.
func (Space) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("content_sources", ContentSource.Type),
	}
}
