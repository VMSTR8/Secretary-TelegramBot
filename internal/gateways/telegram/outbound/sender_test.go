package outbound

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSanitizeHTML verifies that sanitizeHTML keeps the tags allowed by
// Telegram HTML parse mode, strips unknown tags while keeping their inner
// text, and escapes stray angle brackets so plain text is never confused
// with markup.
func TestSanitizeHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain text without tags is returned as is",
			in:   "Привет, мир!",
			want: "Привет, мир!",
		},
		{
			name: "preserves italic tag from llm output",
			in:   `Привет. <i>шёпот</i> мира.`,
			want: `Привет. <i>шёпот</i> мира.`,
		},
		{
			name: "preserves bold tag",
			in:   `<b>жирный</b>`,
			want: `<b>жирный</b>`,
		},
		{
			name: "keeps nested allowed tags",
			in:   `<b>жирный <i>и курсив</i></b>`,
			want: `<b>жирный <i>и курсив</i></b>`,
		},
		{
			name: "preserves all telegram html whitelist tags",
			in: `<b>1</b><strong>2</strong><i>3</i><em>4</em>` +
				`<u>5</u><ins>6</ins><s>7</s><strike>8</strike><del>9</del>` +
				`<code>10</code><pre>11</pre>` +
				`<tg-spoiler>12</tg-spoiler><blockquote>13</blockquote>`,
			want: `<b>1</b><strong>2</strong><i>3</i><em>4</em>` +
				`<u>5</u><ins>6</ins><s>7</s><strike>8</strike><del>9</del>` +
				`<code>10</code><pre>11</pre>` +
				`<tg-spoiler>12</tg-spoiler><blockquote>13</blockquote>`,
		},
		{
			name: "strips disallowed tag but keeps inner text",
			in:   `<script>alert(1)</script>хвост`,
			want: `alert(1)хвост`,
		},
		{
			name: "strips div wrapper, keeps inner allowed markup",
			in:   `<div><i>курсив</i></div>`,
			want: `<i>курсив</i>`,
		},
		{
			name: "escapes lone less-than and greater-than in plain text",
			in:   `5 < 7 и a > b`,
			want: `5 &lt; 7 и a &gt; b`,
		},
		{
			name: "escapes double quotes inside text",
			in:   `он сказал "привет"`,
			want: `он сказал &#34;привет&#34;`,
		},
		{
			name: "keeps anchor href, drops other attributes",
			in:   `<a href="https://example.com" onclick="x()">ссылка</a>`,
			want: `<a href="https://example.com">ссылка</a>`,
		},
		{
			name: "escapes javascript scheme inside href",
			in:   `<a href="javascript:alert(1)">click</a>`,
			want: `<a href="javascript:alert(1)">click</a>`,
		},
		{
			name: "multiline noir text with italic stage directions",
			in: "Дождь стекает по стеклу. <i>Щелчок зажигалки.</i> Город молчит.\n\n" +
				"<i>Выдыхает дым.</i> И снова тишина.",
			want: "Дождь стекает по стеклу. <i>Щелчок зажигалки.</i> Город молчит.\n\n" +
				"<i>Выдыхает дым.</i> И снова тишина.",
		},
		{
			name: "unclosed italic tag is auto-closed by parser",
			in:   `<i>забыл закрыть`,
			want: `<i>забыл закрыть</i>`,
		},
		{
			name: "ampersand is escaped",
			in:   `Tom & Jerry`,
			want: `Tom &amp; Jerry`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, sanitizeHTML(tc.in))
		})
	}
}
