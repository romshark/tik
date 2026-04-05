# FAQ

## What's the benefit of TIK over simple key-based translation?

Abstract keys like `dashboard.newsfeed.summary` hide the actual message from the developer or AI agent reading the code. Human language is highly ambiguous and context-dependent, and an opaque key offers both developers and translators no signal about how the message is actually phrased, used, or intended. TIK embeds the message directly in the source as the key itself.

## Can IDE plugins preload translation texts by key as an alternative?

Preloading reveals the message in the IDE but does not enable static analysis or pipeline automation, depends on a plugin compatible with the specific translation file format and message encoding (ICU, Fluent, ARB), and breaks down entirely outside the IDE (e.g. when browsing code on GitHub).

## Could Fluent be used instead of ICU?

[Fluent](https://projectfluent.org/) is a newer alternative to [ICU MessageFormat](https://github.com/projectfluent/fluent/wiki/Fluent-and-ICU-MessageFormat) but ICU was selected due to wider adoption.
