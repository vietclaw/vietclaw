package tools_test

import (
	"context"
	"testing"
	"vietclaw/internal/tools"
)

func TestParseDDGHTML(t *testing.T) {
	mockHTML := `
	<html>
	<body>
		<div class="result results_links results_links_deep web-result ">
			<div class="links_main links_deep result__body">
				<h2 class="result__title">
					<a class="result__a" rel="nofollow" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com%2Fpage1&amp;rut=1">Example Page 1</a>
				</h2>
				<a class="result__url" href="https://example.com/page1">example.com/page1</a>
				<a class="result__snippet" href="https://example.com/page1">This is the snippet for page 1. It contains some text.</a>
			</div>
		</div>
		<div class="result results_links results_links_deep web-result ">
			<div class="links_main links_deep result__body">
				<h2 class="result__title">
					<a class="result__a" rel="nofollow" href="/l/?uddg=https%3A%2F%2Fexample.com%2Fpage2">Example Page 2</a>
				</h2>
				<a class="result__snippet" href="https://example.com/page2">Snippet for page 2.</a>
			</div>
		</div>
	</body>
	</html>
	`

	results := tools.ParseDDGHTML(mockHTML)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Title != "Example Page 1" {
		t.Errorf("expected title 'Example Page 1', got '%s'", results[0].Title)
	}
	if results[0].URL != "https://example.com/page1" {
		t.Errorf("expected url 'https://example.com/page1', got '%s'", results[0].URL)
	}
	if results[0].Snippet != "This is the snippet for page 1. It contains some text." {
		t.Errorf("unexpected snippet for results[0]: '%s'", results[0].Snippet)
	}

	if results[1].Title != "Example Page 2" {
		t.Errorf("expected title 'Example Page 2', got '%s'", results[1].Title)
	}
	if results[1].URL != "https://example.com/page2" {
		t.Errorf("expected url 'https://example.com/page2', got '%s'", results[1].URL)
	}
	if results[1].Snippet != "Snippet for page 2." {
		t.Errorf("unexpected snippet for results[1]: '%s'", results[1].Snippet)
	}
}

func TestStripHTML(t *testing.T) {
	htmlInput := `
	<html>
		<head>
			<style>
				body { color: red; }
			</style>
			<script>
				console.log("hello");
			</script>
		</head>
		<body>
			<h1>Hello World</h1>
			<p>This is a <b>simple</b> paragraph.</p>
		</body>
	</html>
	`

	expected := "Hello World\nThis is a simple paragraph."
	got := tools.StripHTML(htmlInput)
	if got != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, got)
	}
}

func TestWebSearchEmptyResultsMessage(t *testing.T) {
	results := tools.ParseDDGHTML("<html><body><p>no results</p></body></html>")
	if len(results) != 0 {
		t.Fatalf("expected empty parse")
	}
}

func TestWebSearchRunValidation(t *testing.T) {
	ws := tools.WebSearch{}
	_, err := ws.Run(context.Background(), `{"query": ""}`)
	if err == nil {
		t.Error("expected error with empty query, got nil")
	}
}

func BenchmarkStripHTML(b *testing.B) {
	htmlInput := `
	<html>
		<head>
			<style>
				body { color: red; margin: 0; padding: 0; font-family: sans-serif; }
				.content { padding: 20px; }
			</style>
			<script>
				console.log("hello");
				function doSomething() { alert('test'); }
			</script>
		</head>
		<body>
			<div class="content">
				<h1>Hello World</h1>
				<p>This is a <b>simple</b> paragraph with <i>some formatting</i>.</p>
				<ul>
					<li>Item 1</li>
					<li>Item 2</li>
				</ul>
				<p>
					Another paragraph with multiple     spaces    and
					newlines to test whitespace normalization.
				</p>
				<div>
					<span>Nested</span> <span>Tags</span>
				</div>
			</div>
		</body>
	</html>
	`

	// Duplicate the input to make it larger for a more realistic benchmark
	for i := 0; i < 5; i++ {
		htmlInput += htmlInput
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tools.StripHTML(htmlInput)
	}
}
