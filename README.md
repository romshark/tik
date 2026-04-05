**Author:** Roman Scharkov <roman.scharkov@gmail.com>;
**Version:** 0.9.0;
**Last updated:** 2026-04-05;

# TIK - Textual Internationalization Key

**Table of Contents**

- [Introduction](#introduction)
- [Problem](#problem)
  - [Key-based Translation](#key-based-translation)
  - [ICU Messages](#icu-messages)
- [TIK Syntax Rules](#tik-syntax-rules)
  - [Context](#context)
    - [Context Invariants](#context-invariants)
    - [Context - Example](#context---example)
  - [Body](#body)
  - [Placeholders](#placeholders)
  - [Cardinal Pluralization](#cardinal-pluralization)
    - [Cardinal Pluralization Invariants](#cardinal-pluralization-invariants)
  - [String Placeholders](#string-placeholders)
    - [String Placeholders with Gender](#string-placeholders-with-gender)
- [ICU Encoding](#icu-encoding)
  - [Positional Argument Mapping](#positional-argument-mapping)
- [Configuration Guidelines](#configuration-guidelines)
  - [Domains](#domains)
- [Limitations](#limitations)
- [Standards and Conventions](#standards-and-conventions)
- [FAQ](#faq)
  - [Is this overcomplication really worth it and aren't simple keys enough?](#is-this-overcomplication-really-worth-it-and-arent-simple-keys-enough)
  - [How about just preloading translation texts by key using IDE plugins?](#how-about-just-preloading-translation-texts-by-key-using-ide-plugins)
  - [Could Fluent be used instead of ICU?](#could-fluent-be-used-instead-of-icu)
- [Special Thanks](#special-thanks)

## Introduction

"TIK" is an abbreviation for "Textual Internationalization Key".
A TIK is simultaneously the source of truth for translation and a unique message
identifier within a domain.

TIKs make translation keys human-readable by closely reflecting the actual text shown
to the end users in the source code. This improves context for translators,
enables programmatic generation of
[ICU messages](https://unicode-org.github.io/icu/userguide/format_parse/messages/),
and supports better automation and CI/CD integration.

TIK enables more efficient workflows by integrating TIK processors with CI and LLMs
to give developers immediate feedback on i18n issues before they hit production.
It reduces costs by minimizing reliance on human translators and eases pressure on them
by offloading routine tasks, allowing experts to focus more on quality assurance.

![TIK i18n workflow](https://github.com/romshark/tik/blob/main/tik-i18n-workflow.svg)

TIK is designed to be agnostic to both programming languages and natural languages
used in application source code.

TIP: Check out the official [TIK cheatsheet](https://romshark.github.io/tik-cheatsheet/).

## Problem

Internationalization (i18n) and localization (l10n) are hard — and most developers
avoid them. Supporting multiple languages and regions demands significant effort,
expensive tooling, complex error-prone workflows with slow feedback loops,
and discipline that many teams are unable to take on.

- Translators often work with vague context, leading to broken translations.
- Messages get over-abstracted for reuse breaking grammar and structure in many languages.
- Automation is limited by missing metadata and pipelines developers lack control over.
- The feedback loop is slow, brittle, and disconnected from CI/CD.

The result is missing or poor i18n and l10n that signals lack of polish, undermines
user trust, alienates global audiences and subsequently blocks adoption and growth.

### Key-based Translation

Traditional internationalization relies heavily on key-based systems, where developers
assign abstract message identifiers (e.g. `"dashboard.report.summary"`) to translated
strings stored in external files.

```go
i18n.ByKey("dashboard.report.summary", numberOfMessages, dateTime)
```

Keys offer clear benefits, such as:

- **Separation of concerns -** Developers reference keys,
  while translators manage the actual text.
- **Reusability** - the same message can be used across different contexts or interfaces.
- **Dynamic updates** - translation changes go live immediately without redeployment.
- **Integration** - keys work seamlessly with most existing localization infrastructure.

However, key-based i18n introduces an abstraction layer between the source code
and the actual text, making it harder for developers to immediately understand what
message is being displayed - and in what form.

Naming is inherently hard - and coming up with meaningful, consistent translation keys
can be difficult, especially at scale. Poorly chosen keys often lead to confusion,
redundancy, or fragile reuse patterns.

TIKs, by contrast, embed the meaning directly in the code using a naturally readable
and self-explanatory format that serves as source of truth for the i18n pipeline:

```go
reader.String(`You had {# messages} at {time-short}.`, numberOfMessages, dateTime)
```

### ICU Messages

[ICU messages](https://support.crowdin.com/icu-message-syntax) are a powerful
internationalization tool but are too complex, unreadable and error-prone when used
directly inside the application source code.

Consider the following example in Go:

```go
i18n.Text(`You had {numberOfMessages, plural,
	=0 {no messages}
	one {# message}
	other {# messages}
} at {time, date, jm}.`, numberOfMessages, dateTime)
```

With TIK, developers write simple, readable keys and still get the full power of
ICU under the hood.

## TIK Syntax Rules

```
[ignored spaces] [optional context [ignored spaces]] [body] [ignored spaces]
```

A TIK consists of an optional context and the required text body while the surrounding
unicode spaces are ignored. Both the context and text body must not be empty.

### Context

The TIK context is an optional namespace used to disambiguate message keys.
It is not part of the message’s text body and hence must not be included in the generated ICU message.
If a TIK starts with an opening square bracket `[` then everything up to the next
closing square bracket `]` is treated as the context.

⚠️ Do not confuse context with the message description.

```go
// description.
reader.String(`[context] Text.`)
```

#### Context Invariants

Curly braces `{` `}`, square brackets `[` `]` and reverse-solidus `\`
are not allowed inside the context:

```
[{invalid} context] Text.
```

```
[[invalid context]] Text.
```

```
[invalid\context] Text.
```

The context must not be empty:

```
[ ] This context is invalid.
```

```
[] This context is invalid.
```

#### Context - Example

TIKs are unique message keys within a domain. The same TIK must not be declared more than
once in the source code of a domain unless every occurrence shares the exact same context.
TIK processors enforce this by raising a build-time error for any redeclaration, so
unintended collisions cannot silently collapse into a single ICU message and lose
translation information.

Human language is ambiguous and context-dependent - the same original message text can
have different meanings depending on usage. In such cases, a distinct context must be
added to disambiguate each occurrence.

Example: the word "Order" can mean a sort order, a purchase-order action, or a customer
order, and many languages require different translations for each meaning:

| Meaning           | German          | French        | Russian        |
| :---------------- | :-------------- | :------------ | :------------- |
| sort order        | `"Reihenfolge"` | `"Ordre"`     | `"Порядок"`    |
| place a purchase  | `"Bestellen"`   | `"Commander"` | `"Заказать"`   |
| a customer order  | `"Bestellung"`  | `"Commande"`  | `"Заказ"`      |

The following source code therefore produces build-time errors because the TIK `Order`
is declared three times in the same domain without a disambiguating context:

```html
<!-- orders_page.html -->
<body>
  <table>
    <thead>
      <tr>
        <th>{ i18n.Text(`Name`) }</th>
        <th>{
          // "order" as a sort-order column header.  <--- HERE
          i18n.Text(`Order`)
        }</th>
      </tr>
    </thead>
    <!-- ... -->
  </table>
  <button>{
    // "order" as in "place a purchase order".  <--- HERE
    i18n.Text(`Order`)
  }</button>
</body>
```

```html
<!-- confirmation_page.html -->
<body>
  <h1>{
    // "order" as in "a customer order".  <--- HERE
    i18n.Text(`Order`)
  }</h1>
  <p>{ i18n.Text(`Dispatched on {date-short}.`) }</p>
</body>
```

Each occurrence must therefore carry its own distinct context:

```go
// "order" as a sort-order column header.
i18n.Text(`[table sort column] Order`)
```

```go
// "order" as in "place a purchase order".
i18n.Text(`[order submission] Order`)
```

```go
// "order" as in "a customer order".
i18n.Text(`[order confirmation] Order`)
```

Each of the resulting TIKs defines a distinct context over the shared text body
`"Order"` and produces a separate ICU message.

Conversely, the same TIK may appear any number of times within a 
domain as long as every occurrence shares the exact same context and
body, since all such occurrences refer to the same message.
For example, the order-submission button may be placed on
both a cart page and a checkout page:

```html
<!-- cart_page.html -->
<button>{ i18n.Text(`[order submission] Order`) }</button>
```

```html
<!-- checkout_page.html -->
<button>{ i18n.Text(`[order submission] Order`) }</button>
```

Both occurrences resolve to the same TIK and therefore to a single shared ICU message.

### Body

The text body must always be written in
[CLDR plural rule `other`](https://cldr.unicode.org/index/cldr-spec/plural-rules).
This allows a TIK to avoid branched statements like ICU plural arguments.

### Placeholders

Placeholders allow TIKs to be easily readable yet auto-translatable to ICU message format.
Below is an example TIK that uses multiple placeholders for different data types:

```
Today {name} earned {currency} for completing {# tasks} in section '{text}' at {time-short}.
```

- `{text}` [Text placeholder](#string-placeholders)
- `{name}` [Text placeholder with gender information](#string-placeholders-with-gender)
- `{integer}` [Integer](#icu-encoding---integer)
- `{number}` [Number](#icu-encoding---number)
- `{# ...}` [Cardinal pluralization](#icu-encoding---cardinal-pluralization)
- `{ordinal}` [Ordinal pluralization](#icu-encoding---ordinal-pluralization)
- `{date-full}` [Date placeholder](#icu-encoding---datetime-placeholders)
- `{date-long}` [Date placeholder](#icu-encoding---datetime-placeholders)
- `{date-medium}` [Date placeholder](#icu-encoding---datetime-placeholders)
- `{date-short}` [Date placeholder](#icu-encoding---datetime-placeholders)
- `{time-full}` [Time placeholder](#icu-encoding---datetime-placeholders)
- `{time-long}` [Time placeholder](#icu-encoding---datetime-placeholders)
- `{time-medium}` [Time placeholder](#icu-encoding---datetime-placeholders)
- `{time-short}` [Time placeholder](#icu-encoding---datetime-placeholders)
- `{currency}` [Currency](#icu-encoding---currency)

### Cardinal Pluralization

A pluralization statement `{# ...}` begins with `{# ` and ends with `}`.
The `#` is the placeholder for the actual number value (if any).
The contents `...` may contain any contents that aren't explicitly forbidden
(see [invariants](#cardinal-pluralization-invariants)).

The contents may contain any number of placeholders:

```
You had {# messages marked as {text} at {time-long}}
```

```
You had {# tasks} assigned at {time-short}.
```

#### Cardinal Pluralization Invariants

1. Plural statements must not begin and end with a Unicode whitespace character.
(as defined by [Unicode](https://unicode.org/charts/collation/chart_Whitespace.html)):

```
This TIK is illegal: {#  <- two spaces here}
```

```
This TIK is illegal: {# space here-> }
```

2. Plural statements cannot be nested:

```
This TIK is illegal: {# first level {# second level}}
```

3. Plural statement contents cannot start with a placeholder:

```
This TIK is illegal: {# {integer}}
```

```
This TIK is illegal: {# {number}}
```

```
This TIK is illegal: {# {currency}}
```

```
This TIK is illegal: {# {date-full}}
```

### String Placeholders

String placeholders `{text}` represent arbitrary text.

```
You joined group {text}.
```

```
All articles from category: {text}.
```

If the identifier at hand has a gender (like a person's name) then consider using
[a string placeholder with gender](#string-placeholders-with-gender) instead because
for gender-aware locales this might affect the grammar.

#### String Placeholders with Gender

String placeholders `{name}` must be infused with gender information.
This placeholder still represents arbitrary strings values but should be used for
names and identifiers to allow correct translation for gender-aware locales.

```go
reader.String(
	`The journey began, {name} had embarked onto the ship.`, // TIK
	tokibundle.String{ Value: "John", Gender: tokibundle.GenderMale },
)
```

TIK doesn't define how gender information is attached to the placeholder.
This is determined by the TIK processor.

ℹ️ Gender may affect grammar in some languages:

| Language  | masculine         | feminine            |
| :-------- | :---------------- | :------------------ |
| Ukrainian | `John готовий`    | `Martha готова`     |
| Italian   | `John è pronto`   | `Martha è pronta`   |
| French    | `John est prêt`   | `Martha est prête`  |
| Spanish   | `John está listo` | `Martha está lista` |
| Russian   | `John готов`      | `Martha готова`     |

The translated ICU message for locale `uk` would be:

```
Розпочалася подорож, {var0_gender, select,
  female { {var0} вирушила на корабель. }
  male { {var0} вирушив на корабель. }
  other { {var0} вирушило на корабель. }
}
```

## ICU Encoding

| TIK placeholder | ICU equivalent                      |
| :-------------- | :---------------------------------- |
| `{text}`        | `{var0}`                            |
| `{name}`        | `{var0, select, other{...}}`        |
| `{number}`      | `{var0, number}`                    |
| `{integer}`     | `{var0, number, integer}`           |
| `{# ...}`       | `{var0, plural, other{# ...}}`      |
| `{ordinal}`     | `{var0, selectordinal, other{#th}}` |
| `{name}`        | `{var0, select, other{...}}`        |
| `{date-full}`   | `{var0, date, full}`                |
| `{date-long}`   | `{var0, date, long}`                |
| `{date-medium}` | `{var0, date, medium}`              |
| `{date-short}`  | `{var0, date, short}`               |
| `{time-full}`   | `{var0, time, full}`                |
| `{time-long}`   | `{var0, time, long}`                |
| `{time-medium}` | `{var0, time, medium}`              |
| `{time-short}`  | `{var0, time, short}`               |
| `{time-short}`  | `{var0, time, short}`               |
| `{currency}`    | `{var0, number, ::currency/auto}`   |

The `...` stands for any content, meaning that the following TIK:

```
{# messages in {# groups}}
```

Encodes to the following ICU:

```
`{var0, plural, other{# messages in {var0, plural, other{# groups}}}}`
```

### Positional Argument Mapping

All placeholders are mapped positionally, meaning that the order of occurrence in the TIK
is the order expected for argument inputs.

```
[report] By {time-short}, {name} received {# emails}.
```

All placeholders use the `var` prefix with a following positional index.

generated ICU:
```
By { var0, time, short }, { var1_gender, select,
  other { {var1} }
} {var1} received {var2, plural,
  one {# email}
  other {# emails}
}. 
```

Usage example in Go:

```go
reader.String(`[report] By {time-short}, {text} received {# emails}.`,
	time.Now(), "Max", len(emailsReceived))
```

## Configuration Guidelines

The TIK specification defines guidelines only
and imposes no strict format or requirements.
The exact configuration format is left entirely to the processor implementation.

### Domains

A domain is the scope within which TIKs must be unique. The mechanism by which
sources are assigned to a domain is left entirely to the TIK processor implementation:
the specification does not mandate a configuration format, a discovery strategy,
or a default boundary. Implementations are free to treat the entire project as a
single domain, derive domains from the directory layout, or accept explicit mappings.

Large-scale projects with many translations can benefit from partitioning TIKs into
multiple domains. One implementation defines domain scopes through a central
configuration file:

```json
{
  "domains": {
    "domain_common": [
      {
        "dir": "/...",
        "description": "Messages use a neutral, product-wide tone."
      }
    ],
    "domain_A": [
      {
        "dir": "/domain_a/...",
        "description": "Customer-facing online storefront. Messages address end customers in a welcoming, persuasive tone."
      },
      {
        "dir": "/templates/domain_a/_"
      }
    ],
    "domain_B": [
      {
        "dir": "/domain_b/...",
        "description": "Internal warehouse and order-management tooling. Messages address operations staff in a precise, operational tone."
      }
    ]
  }
}
```

Another implementation might derive domains from marker files placed in the
source tree. For example, a `.tikdomain` file could mark a domain boundary that
covers its containing directory and all descendants, with nested `.tikdomain`
files establishing child domains whose descriptions append onto the parent's:

```
repo/
├── .tikdomain
├── domain_a/
│   ├── .tikdomain
│   └── page.html
├── domain_b/
│   ├── .tikdomain
│   └── page.html
└── shared/
    └── widget.html
```

**repo/.tikdomain**:

```
Messages use a neutral, product-wide tone.
```

**repo/domain_a/.tikdomain**:

```
Customer-facing online storefront.
Messages address end customers in a welcoming, persuasive tone.
```

**repo/domain_b/.tikdomain**:

```
Internal warehouse and order-management tooling.
Messages address operations staff in a precise, operational tone.
```

Beyond enforcing TIK uniqueness, domains may carry a human-readable description
that characterizes the scope, audience, or tone of the messages they contain.
Processors can forward these descriptions to LLM or human translators as
additional context, improving the accuracy of translations that would otherwise
be produced from the TIK alone.

## Limitations

As with any technology, TIK introduces both advantages and trade-offs.

- Advantages
  - ✅ **Readability**: TIK keys convey the intent of the message
    in a clear and human-readable format.
  - ✅ **Automation**: The TIK syntax can be programmatically converted into
    ICU message structures and translation can mostly be automated through LLMs.
  - ✅ **Simplicity**: The format is relatively straightforward to learn
    and apply consistently.
- Limitations
  - ⚠️ **Learning Curve**: Developers must become familiar
    with the TIK syntax conventions.
  - ⚠️ **Developer Responsibility**: Developers must write somewhat meaningful texts and
    can't fully rely on translators. They can only rely on translators and software
    to improve those texts later in the translation pipeline.
  - ⚠️ **Tooling Requirements**: A dedicated extractor tool (referred to as TIK processor
    through this document) is required to parse and process TIK keys to eventually
    produce ICU messages for translation.
  - ⚠️ **Source Language Translation**: Messages written in the source language
    (e.g., English) must also be extracted and passed through the translation pipeline.
  - ⚠️ **Static Analysis Scope**: Duplicate-TIK detection is limited to what the
    processor can see in the source code. If a TIK is wrapped in a helper function
    and reused across call sites, the processor observes only a single declaration
    and cannot warn about unintended reuse of the returned message in conflicting
    contexts.

## Standards and Conventions

- Plural categories follow [Unicode CLDR](https://cldr.unicode.org/index/cldr-spec/plural-rules)
- Language codes follow [ISO 639-1](https://www.iso.org/iso-639-language-codes.html)
- Currency codes follow [ISO 4217](https://www.iso.org/iso-4217-currency-codes.html)
- Timestamps follow [RFC3339](https://www.rfc-editor.org/rfc/rfc3339.html)
- JSON examples follow [RFC8259](https://datatracker.ietf.org/doc/html/rfc8259)
- Date/Time [RFC1123](https://datatracker.ietf.org/doc/html/rfc1123)

## FAQ

### Is this overcomplication really worth it and aren't simple keys enough?

The answer depends on perspective. While abstract keys like `dashboard.newsfeed.summary`
offer [clear benefits](#key-based-translation)
they also come with certain [limitations](#problem).
It is likely that, for the foreseeable future, code will continue to be written and
maintained primarily by humans. At the same time, large language models are demonstrating
increasing proficiency in translation tasks. The concept behind TIK is to define clear,
human-readable messages directly in the source code, delegating the complexity of
generating accurate ICU messages for various languages to large language models.

To give you some context, only the last sentence of this answer was actually written
by a human.

### How about just preloading translation texts by key using IDE plugins?

This doesn't address pipeline automation issues but is a theoretically
viable solution to opaque abstract keys in source code. However,  this approach is
inherently limited to IDEs that support such a feature.
Additionally, those IDEs/extensions must be compatible with your
specific translation file format and message encoding (e.g., ICU, Fluent, ARB).
It also breaks down entirely when browsing code outside the IDE - for example,
on GitHub - where no plugin can preload or resolve translation keys.

### Could Fluent be used instead of ICU?

[Fluent](https://projectfluent.org/) can be considered a worthy
[counterpart to the ICU MessageFormat](https://github.com/projectfluent/fluent/wiki/Fluent-and-ICU-MessageFormat)
and technically nothing speaks against using it as an alternative TIK backend
yet ICU was selected due to wider adoption.

## Special Thanks

Special thanks to Muthu Kumar ([@MKRhere](https://github.com/MKRhere))!
