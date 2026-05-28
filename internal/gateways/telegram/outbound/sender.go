package outbound

import (
	"context"
	"fmt"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
	"strings"

	tgmodel "github.com/go-telegram/bot/models"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/go-telegram/bot"
)

var _ repository.BusinessSender = (*Sender)(nil)

// Telegram HTML whitelist: https://core.telegram.org/bots/api#html-style
var tgAllowedTags = map[string]bool{
	"b": true, "strong": true,
	"i": true, "em": true,
	"u": true, "ins": true,
	"s": true, "strike": true, "del": true,
	"code": true, "pre": true,
	"a":          true,
	"tg-spoiler": true,
	"blockquote": true,
}

type Sender struct {
	bot *bot.Bot
}

func NewSender(b *bot.Bot) *Sender {
	return &Sender{
		bot: b,
	}
}

func (s *Sender) ShowThinking(ctx context.Context, target model.ReplyDraft) error {
	_, err := s.bot.SendChatAction(ctx, &bot.SendChatActionParams{
		BusinessConnectionID: target.BusinessConnectionID,
		ChatID:               target.GuestID,
		Action:               tgmodel.ChatActionTyping,
	})
	if err != nil {
		return fmt.Errorf("telegram send chat action: %w", err)
	}

	return nil
}

func (s *Sender) Send(ctx context.Context, draft model.ReplyDraft) error {
	_, err := s.bot.SendMessage(ctx, &bot.SendMessageParams{
		BusinessConnectionID: draft.BusinessConnectionID,
		ChatID:               draft.GuestID,
		Text:                 sanitizeHTML(draft.Text),
		ParseMode:            tgmodel.ParseModeHTML,
	})
	if err != nil {
		return fmt.Errorf("telegram send message: %w", err)
	}

	return nil
}

func sanitizeHTML(text string) string {
	nodes, err := html.ParseFragment(strings.NewReader(text), &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Body,
		Data:     "body",
	})
	if err != nil {
		return html.EscapeString(text)
	}

	var buf strings.Builder
	for _, n := range nodes {
		renderTGNode(&buf, n)
	}

	return buf.String()
}

func renderTGNode(buf *strings.Builder, n *html.Node) {
	switch n.Type {
	case html.TextNode:
		buf.WriteString(html.EscapeString(n.Data))
	case html.ElementNode:
		renderTGElement(buf, n)
	case html.ErrorNode, html.DocumentNode, html.CommentNode, html.DoctypeNode, html.RawNode:
		// ignore
	}
}

func renderTGElement(buf *strings.Builder, n *html.Node) {
	if !tgAllowedTags[n.Data] {
		renderTGChildren(buf, n)

		return
	}

	buf.WriteByte('<')
	buf.WriteString(n.Data)

	if n.Data == "a" {
		writeTGHref(buf, n)
	}

	buf.WriteByte('>')
	renderTGChildren(buf, n)
	buf.WriteString("</")
	buf.WriteString(n.Data)
	buf.WriteByte('>')
}

func writeTGHref(buf *strings.Builder, n *html.Node) {
	for _, a := range n.Attr {
		if a.Key == "href" {
			buf.WriteString(` href="`)
			buf.WriteString(html.EscapeString(a.Val))
			buf.WriteByte('"')
		}
	}
}

func renderTGChildren(buf *strings.Builder, n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderTGNode(buf, c)
	}
}
