package audit

import "testing"

func TestBuildFieldQuery_BasicChinesePhrase(t *testing.T) {
	got := BuildFieldQuery("收到 UpdateMemory_2 请求", "message")
	want := `message:#"收到 UpdateMemory_2 请求"`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildFieldQuery_WithQuotes(t *testing.T) {
	got := BuildFieldQuery("event=\"收到 UpdateMemory_2 请求\"", "message")
	want := `message:#"event=\"收到 UpdateMemory_2 请求\""`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildFieldQuery_WithExistingEscapedQuote_ShouldNotDoubleEscape(t *testing.T) {
	// input already contains escaped quotes (\")
	got := BuildFieldQuery("abc\\\"def\\\"ghi", "msg")
	// all backslashes are escaped and quotes are escaped
	want := `msg:#"abc\\\"def\\\"ghi"`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildFieldQuery_WithStandaloneBackslash(t *testing.T) {
	got := BuildFieldQuery(`path \\ server`, "caller")
	// expect four backslashes in the final string
	want := `caller:#"path \\\\ server"`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildFullTextQuery(t *testing.T) {
	got := BuildFullTextQuery("some: 值 \"quoted\" and \\ slash")
	want := `#"some: 值 \"quoted\" and \\ slash"`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
