# Slack Foli Stop

An app that integrates [Foli](https://www.foli.fi) (local public transport) realtime information on bus departures at stops to Slack.

The app is written in [Go](https://golang.org) and is deployed and executed to [Heroku](https://heroku.com).

Usage: `/<slash command> <stop id>`

# Setup

1. Deploy your application to Heroku
2. Create a Slack [slash command](https://api.slack.com/slash-commands).
3. Set your Heroku application's endpoint to the slash command's url.
3. Set the Slack token to your app's environment variable: `heroku config:set SECRET=<token>`