package html

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/buffer"
)

func assertTokens(t *testing.T, s string, tokentypes ...TokenType) {
	stringify := helperStringify(t, s)
	l := NewLexer(bytes.NewBufferString(s))
	assert.True(t, l.IsEOF(), "lexer must have buffer fully in memory in "+stringify)
	i := 0
	for {
		tt, _ := l.Next()
		if tt == ErrorToken {
			assert.Equal(t, io.EOF, l.Err(), "error must be EOF in "+stringify)
			assert.Equal(t, len(tokentypes), i, "when error occurred we must be at the end in "+stringify)
			break
		}
		assert.False(t, i >= len(tokentypes), "index must not exceed tokentypes size in "+stringify)
		if i < len(tokentypes) {
			assert.Equal(t, tokentypes[i], tt, "tokentypes must match at index "+strconv.Itoa(i)+" in "+stringify)
		}
		i++
	}
	return
}

func assertTags(t *testing.T, s string, tags ...string) {
	stringify := helperStringify(t, s)
	l := NewLexer(bytes.NewBufferString(s))
	i := 0
	for {
		tt, data := l.Next()
		if tt == ErrorToken {
			assert.Equal(t, io.EOF, l.Err(), "error must be EOF in "+stringify)
			assert.Equal(t, len(tags), i, "when error occurred we must be at the end in "+stringify)
			break
		} else if tt == StartTagToken || tt == EndTagToken || tt == DoctypeToken {
			assert.False(t, i >= len(tags), "index must not exceed tags size in "+stringify)
			if i < len(tags) {
				assert.Equal(t, tags[i], string(data), "tags must match at index "+strconv.Itoa(i)+" in "+stringify)
				i++
			}
		}
	}
	return
}

func assertAttributes(t *testing.T, s string, attributes ...string) {
	stringify := helperStringify(t, s)
	l := NewLexer(bytes.NewBufferString(s))
	i := 0
	for {
		tt, data := l.Next()
		if tt == ErrorToken {
			assert.Equal(t, io.EOF, l.Err(), "error must be EOF in "+stringify)
			assert.Equal(t, len(attributes), i, "when error occurred we must be at the end in "+stringify)
			break
		} else if tt == AttributeToken {
			assert.False(t, i+1 >= len(attributes), "index must not exceed attributes size in "+stringify)
			if i+1 < len(attributes) {
				assert.Equal(t, attributes[i], string(data), "attribute keys must match at index "+strconv.Itoa(i)+" in "+stringify)
				assert.Equal(t, attributes[i+1], string(l.AttrVal()), "attribute values must match at index "+strconv.Itoa(i)+" in "+stringify)
				i += 2
			}
		}
	}
	return
}

func helperStringify(t *testing.T, input string) string {
	s := ""
	l := NewLexer(bytes.NewBufferString(input))
	for i := 0; i < 10; i++ {
		tt, text := l.Next()
		if tt == ErrorToken {
			s += tt.String() + "('" + l.Err().Error() + "')"
			break
		} else if tt == AttributeToken {
			s += tt.String() + "('" + string(text) + "=" + string(l.AttrVal()) + "') "
		} else {
			s += tt.String() + "('" + string(text) + "') "
		}
	}
	return s
}

////////////////////////////////////////////////////////////////

func TestTokens(t *testing.T) {
	assertTokens(t, "<html></html>", StartTagToken, StartTagCloseToken, EndTagToken)
	assertTokens(t, "<img/>", StartTagToken, StartTagVoidToken)
	assertTokens(t, "<!-- comment -->", CommentToken)
	assertTokens(t, "<!-- comment --!>", CommentToken)
	assertTokens(t, "<p>text</p>", StartTagToken, StartTagCloseToken, TextToken, EndTagToken)
	assertTokens(t, "<input type='button'/>", StartTagToken, AttributeToken, StartTagVoidToken)
	assertTokens(t, "<input  type='button'  value=''/>", StartTagToken, AttributeToken, AttributeToken, StartTagVoidToken)
	assertTokens(t, "<input type='=/>' \r\n\t\f value=\"'\" name=x checked />", StartTagToken, AttributeToken, AttributeToken, AttributeToken, AttributeToken, StartTagVoidToken)
	assertTokens(t, "<!doctype>", DoctypeToken)
	assertTokens(t, "<!doctype html>", DoctypeToken)
	assertTokens(t, "<?bogus>", CommentToken)
	assertTokens(t, "</0bogus>", CommentToken)
	assertTokens(t, "<!bogus>", CommentToken)
	assertTokens(t, "< ", TextToken)
	assertTokens(t, "</", TextToken)

	// raw tags
	assertTokens(t, "<title><p></p></title>", StartTagToken, StartTagCloseToken, TextToken, EndTagToken)
	assertTokens(t, "<TITLE><p></p></TITLE>", StartTagToken, StartTagCloseToken, TextToken, EndTagToken)
	assertTokens(t, "<plaintext></plaintext>", StartTagToken, StartTagCloseToken, TextToken)
	assertTokens(t, "<script></script>", StartTagToken, StartTagCloseToken, EndTagToken)
	assertTokens(t, "<script>var x='</script>';</script>", StartTagToken, StartTagCloseToken, TextToken, EndTagToken, TextToken, EndTagToken)
	assertTokens(t, "<script><!--var x='</script>';--></script>", StartTagToken, StartTagCloseToken, TextToken, EndTagToken, TextToken, EndTagToken)
	assertTokens(t, "<script><!--var x='<script></script>';--></script>", StartTagToken, StartTagCloseToken, TextToken, EndTagToken)
	assertTokens(t, "<script><!--var x='<script>';--></script>", StartTagToken, StartTagCloseToken, TextToken, EndTagToken)
	assertTokens(t, "<![CDATA[ test ]]>", TextToken)

	// early endings
	assertTokens(t, "<!-- comment", CommentToken)
	assertTokens(t, "<foo", StartTagToken)
	assertTokens(t, "</foo", EndTagToken)
	assertTokens(t, "<foo x", StartTagToken, AttributeToken)
	assertTokens(t, "<foo x=", StartTagToken, AttributeToken)
	assertTokens(t, "<foo x='", StartTagToken, AttributeToken)
	assertTokens(t, "<foo x=''", StartTagToken, AttributeToken)
	assertTokens(t, "<!DOCTYPE note SYSTEM", DoctypeToken)
	assertTokens(t, "<![CDATA[ test", TextToken)
	assertTokens(t, "<script>", StartTagToken, StartTagCloseToken)
	assertTokens(t, "<script><!--", StartTagToken, StartTagCloseToken, TextToken)
	assertTokens(t, "<script><!--var x='<script></script>';-->", StartTagToken, StartTagCloseToken, TextToken)

	buffer.MinBuf = 4
	assert.Equal(t, "StartTag('ab') StartTagClose('>') Error('EOF')", helperStringify(t, "<ab   >"), "buffer reallocation must keep tagname valid")

	assert.Equal(t, "Invalid(100)", TokenType(100).String())
}

func TestTags(t *testing.T) {
	assertTags(t, "<foo:bar.qux-norf/>", "foo:bar.qux-norf")
	assertTags(t, "<foo?bar/qux>", "foo?bar/qux")
	assertTags(t, "<!DOCTYPE note SYSTEM \"Note.dtd\">", "note SYSTEM \"Note.dtd\"")

	// early endings
	assertTags(t, "<foo ", "foo")
}

func TestAttributes(t *testing.T) {
	assertAttributes(t, "<foo a=\"b\" />", "a", "\"b\"")
	assertAttributes(t, "<foo \nchecked \r\n value\r=\t'=/>\"' />", "checked", "", "value", "'=/>\"'")
	assertAttributes(t, "<foo bar=\" a \n\t\r b \" />", "bar", "\" a \n\t\r b \"")
	assertAttributes(t, "<foo a/>", "a", "")
	assertAttributes(t, "<foo /=/>", "/", "/")

	// early endings
	assertAttributes(t, "<foo x", "x", "")
	assertAttributes(t, "<foo x=", "x", "")
	assertAttributes(t, "<foo x='", "x", "'")
}

////////////////////////////////////////////////////////////////

func ExampleNewLexer() {
	l := NewLexer(bytes.NewBufferString("<span class='user'>John Doe</span>"))
	out := ""
	for {
		tt, data := l.Next()
		if tt == ErrorToken {
			break
		}
		if tt == StartTagToken {
			out += "<"
		} else if tt == EndTagToken {
			out += "</"
		}
		out += string(data)
		if tt == StartTagToken {
			out += " "
		} else if tt == EndTagToken {
			out += ">"
		} else if tt == AttributeToken {
			out += "=" + string(l.AttrVal())
		}
	}
	fmt.Println(out)
	// Output: <span class='user'>John Doe</span>
}
