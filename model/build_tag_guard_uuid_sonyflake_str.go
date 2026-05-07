//go:build uuid && sonyflake_str

package model

// Trigger a readable compile-time failure when mutually-exclusive tags are enabled together.
var _ = uuid_and_sonyflake_str_build_tags_are_mutually_exclusive
