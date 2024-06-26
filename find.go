package sinonimos

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gosimple/slug"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
//	"golang.org/x/text/encoding/charmap"
)

var (
	// ErrNotFound is returned when an expression is not found on sinonimos.com.br.
	ErrNotFound          = errors.New("expression not found")
	// ErrHTTPLayer is returned when internet connection is not available.
	ErrHTTPLayer         = errors.New("an error launched when trying to access the website")
	// ErrInvalidFormatBody is returned when body from HTML response is not valid to parse.
	ErrInvalidFormatBody = errors.New("it was not possible to parse the received html")
)

// Meaning contains information about an meaning.
//
// See Also
//
// Find
type Meaning struct {
	Description string
	Synonyms    []string
	Examples    []string
}

// FindInput contains the input data require to Find.
//
// See Also
//
// Find
type FindInput struct {
	Expression string
}

// FindOutput contains the output payload from Find.
//
// See Also
//
// Find
type FindOutput struct {
	Meanings []Meaning
}

// Find try to find meanings for an expression on sinonimos.com.br.
func Find(input *FindInput) (*FindOutput, error) {
	resp, err := http.Get(fmt.Sprintf("https://www.sinonimos.com.br/%s/", slug.Make(input.Expression)))
	if err != nil {
		return nil, ErrHTTPLayer
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	root, err := html.Parse(resp.Body)
	resp.Body.Close() // Close the response body after parsing
	if err != nil {
		return nil, ErrInvalidFormatBody
	}

	meaningSections := scrape.FindAll(root, scrape.ByClass("content-detail"))
	meanings := make([]Meaning, len(meaningSections))

	for j, meaningSection := range meaningSections {
		if meaning, ok := scrape.Find(meaningSection, scrape.ByClass("syn-list")); ok {
			meanings[j].Description = strings.TrimSpace(scrape.Text(meaning))
		}

		synonyms := scrape.FindAll(meaningSection, synonymMatcher)
		meanings[j].Synonyms = make([]string, len(synonyms))
		for i, synonym := range synonyms {
			meanings[j].Synonyms[i] = strings.TrimSpace(scrape.Text(synonym))
		}

		examples := scrape.FindAll(meaningSection, scrape.ByClass("content-info"))
		meanings[j].Examples = make([]string, len(examples))
		for i, example := range examples {
			meanings[j].Examples[i] = strings.TrimSpace(scrape.Text(example))
		}
	}

	return &FindOutput{
		Meanings: meanings,
	}, nil
}

func synonymMatcher(n *html.Node) bool {
	if n.DataAtom == atom.A || n.DataAtom == atom.Span {
		if n.Parent != nil {
			return scrape.Attr(n.Parent, "class") == "sinonimos" && scrape.Attr(n, "class") != "exemplo"
		}
	}
	return false
}
