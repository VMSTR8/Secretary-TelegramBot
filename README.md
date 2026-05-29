[Русская версия](README_RU.md)
## A secretary bot for personal messages…

…that really, really doesn't like it when somebody writes:

❌ Hi!

❌ Yo

❌ Good afternoon

If you're here, you either bumped into my little project by accident, or you're also fed up with folks who say "hi" 
first and only then start typing what they actually want. Or, even better, sit and wait for your reaction to their 
"hi!".

And then there are people who write
```
one short message 

at a time constantly 

producing some kind 

of flood wall
```
Both kinds are present in my social circle. I don't enjoy being rude, and not everyone would get why such trivia gets 
on my nerves. So I wrote a bot!

## What does your smart machine actually do?

It looks at a blacklist of words from the configuration. If the other party writes any of them — with or without 
punctuation, doesn't matter — the bot reacts. It hits an LLM (DeepSeek in my case, since it's pretty cheap), 
uses your configured prompt to generate a reply, and sends it back to the person for you.

On top of that, if someone is flooding, the bot will also reply on every 5th message within the time window specified 
in the config.

The wording of the reply depends on the prompt you configure.

Telegram has opened the ability to attach a secretary bot to any account, even without Premium. Business folks are 
happy about this feature, and ~~dork~~ special programmers like me look for ways to have a bit of fun with it.

## TL;DR
The bot reacts to a single greeting or to flooding within a defined interval window.

## Stack

- Go 1.26.3
- [Golang Telegram Bot](https://github.com/go-telegram/bot)
- [DeepSeek API](https://platform.deepseek.com/)

Redis or some tiny DB like SQLite is on the TODO list, in case the bot needs to grow.

## How to deploy on your own VDS

TBA