package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// ContentSource holds the schema definition for the ContentSource entity.
type ContentSource struct {
	ent.Schema
}

// Fields of the ContentSource.
func (ContentSource) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("space_id", uuid.UUID{}),
		field.UUID("owner_id", uuid.UUID{}),
		field.String("title"),
		field.String("media_type"), // e.g., document, audio, video
		field.String("source"),     // e.g., upload, gdrive, paste
		field.String("status").
			Default("UPLOADING"),
		field.String("original_blob_hash").
			MaxLen(64),
		field.String("processed_blob_hash").
			MaxLen(64).
			Optional().
			Nillable(),
		field.Text("content_summary").
			Optional().
			Nillable(),
		field.Strings("keywords").
			Optional(),
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

// Edges of the ContentSource.
func (ContentSource) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("space", Space.Type).
			Ref("content_sources").
			Field("space_id").
			Unique().
			Required(),
	}
}
