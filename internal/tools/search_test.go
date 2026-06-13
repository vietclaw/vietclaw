package tools

import (
	"testing"
)

var dummyDDGHTML = `<!DOCTYPE html><html><body>
<div class="result result--html">
	<div class="result__body">
		<a class="result__a" href="https://example.com/1?uddg=https%3A%2F%2Fexample.com%2Freal-url-1&amp;rut=1">
			Example <b>Domain</b> 1
		</a>
		<a class="result__snippet" href="#">This is a snippet for example 1.</a>
	</div>
</div>
<div class="result result--html">
	<div class="result__body">
		<a class="result__a" href="https://example.com/2?uddg=https%3A%2F%2Fexample.com%2Freal-url-2&amp;rut=1">
			Example <b>Domain</b> 2
		</a>
		<a class="result__snippet" href="#">This is a snippet for example 2.</a>
	</div>
</div>
<div class="result result--html">
	<div class="result__body">
		<a class="result__a" href="https://example.com/3?uddg=https%3A%2F%2Fexample.com%2Freal-url-3&amp;rut=1">
			Example <b>Domain</b> 3
		</a>
		<a class="result__snippet" href="#">This is a snippet for example 3.</a>
	</div>
</div>
</body></html>`

func BenchmarkParseDDGHTML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDDGHTML(dummyDDGHTML)
	}
}

func TestParseDDGHTML(t *testing.T) {
	results := ParseDDGHTML(dummyDDGHTML)
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	if results[0].Title != "Example Domain 1" {
		t.Errorf("Expected title 'Example Domain 1', got '%s'", results[0].Title)
	}
	if results[0].URL != "https://example.com/real-url-1" {
		t.Errorf("Expected URL 'https://example.com/real-url-1', got '%s'", results[0].URL)
	}
}
