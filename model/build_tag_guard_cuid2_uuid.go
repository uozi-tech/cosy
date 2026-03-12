//go:build cuid2 && uuid

package model

// Trigger a readable compile-time failure when mutually-exclusive tags are enabled together.
var _ = cuid2_and_uuid_build_tags_are_mutually_exclusive
