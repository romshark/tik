**Author:** Roman Scharkov <roman.scharkov@gmail.com>;
**Version:** 0.9.0;
**Last updated:** 2026-04-05;

# TIK Specification

This document is the normative specification for the Textual Internationalization Key (TIK) format.

**Table of Contents**

- [TIK Syntax Rules](#tik-syntax-rules)
  - [Context](#context)
    - [Context - Syntactic Invariants](#context---syntactic-invariants)
    - [Context Uniqueness](#context-uniqueness)
    - [Context - Example](#context---example)
  - [Body](#body)
  - [Placeholders](#placeholders)
  - [Cardinal Pluralization](#cardinal-pluralization)
    - [Cardinal Pluralization - Blank Pluralization Example](#cardinal-pluralization---blank-pluralization-example)
    - [Cardinal Pluralization - Syntactic Invariants](#cardinal-pluralization---syntactic-invariants)
  - [String Placeholders](#string-placeholders)
    - [String Placeholders with Gender](#string-placeholders-with-gender)
- [ICU Encoding](#icu-encoding)
  - [Positional Argument Mapping](#positional-argument-mapping)
- [Configuration Guidelines](#configuration-guidelines)
  - [Domains](#domains)
- [Standards and Conventions](#standards-and-conventions)

## TIK Syntax Rules

```
[ignored whitespace] [[context] whitespace] [body] [ignored whitespace]
```

A TIK consists of an optional context followed by a required text body. Unicode whitespace surrounding the TIK and trailing the body is ignored. A component is **empty** when it contains no characters other than Unicode whitespace. When a context is present, it must not be empty, and at least one Unicode whitespace character must separate the closing `]` of the context from the body. The body must never be empty.

### Context

The TIK context is an optional namespace used to disambiguate message keys. It is not part of the message’s text body and hence must not be included in the generated ICU message. If a TIK starts with an opening square bracket `[`, then everything up to the next closing square bracket `]` is treated as the context. If no closing `]` is found, the TIK is invalid.

The TIK context is distinct from the message description and is not interchangeable with it.

```go
// description.
reader.String(`[context] Text.`)
```

#### Context - Syntactic Invariants

Unlike the [body](#body), the context does not support escape sequences. Curly braces `{` `}`, square brackets `[` `]` and reverse-solidus `\` are not allowed inside the context:

```
[{invalid} context] Text.
```

```
[[invalid context]] Text.
```

```
[invalid\context] Text.
```

An opening `[` without a matching `]` is invalid:

```
[unclosed context Text.
```

The context must be followed by at least one whitespace character before the body:

```
[context]Text without separator.
```

The context must not be empty:

```
[ ] This context is invalid.
```

```
[] This context is invalid.
```

#### Context Uniqueness

A TIK without a [context](#context) must not be declared more than once in the source code of a [domain](#domains). A TIK with a context may appear multiple times within the same domain as long as every occurrence shares the exact same context and body, in which case all occurrences resolve to a single shared ICU message. TIK processors enforce these rules by raising a build-time error for any violation.

#### Context - Example

Human language is ambiguous and context-dependent - the same original message text can have different meanings depending on usage. In such cases, a distinct context must be added to disambiguate each occurrence.

The word "Order" can mean a sort order, a purchase-order action, or a customer order, and many languages require different translations for each meaning:

| Meaning           | German          | French        | Russian        |
| :---------------- | :-------------- | :------------ | :------------- |
| sort order        | `"Reihenfolge"` | `"Ordre"`     | `"Порядок"`    |
| place a purchase  | `"Bestellen"`   | `"Commander"` | `"Заказать"`   |
| a customer order  | `"Bestellung"`  | `"Commande"`  | `"Заказ"`      |

The following source code produces build-time errors because the TIK `Order` is declared three times in the same domain without a disambiguating context:

```html
<!-- orders_page.html -->
<body>
  <table>
    <thead>
      <tr>
        <th>{ i18n.Text(`Name`) }</th>
        <th>{ i18n.Text(`Order`) }</th>       <!-- HERE -->
      </tr>
    </thead>
  </table>
  <button>{ i18n.Text(`Order`) }</button>     <!-- HERE -->
</body>
```

```html
<!-- confirmation_page.html -->
<body>
  <h1>{ i18n.Text(`Order`) }</h1>             <!-- HERE -->
  <p>{ i18n.Text(`Dispatched on {date-short}.`) }</p>
</body>
```

Each occurrence must carry its own distinct context:

```
[table sort column] Order
[order submission] Order
[order confirmation] Order
```

Each resulting TIK produces a separate ICU message.

Conversely, a TIK with context may be reused across files. Both pages below share a single ICU message because the context and body are identical:

```html
<!-- cart_page.html -->
<button>{ i18n.Text(`[order submission] Order`) }</button>
```

```html
<!-- checkout_page.html -->
<button>{ i18n.Text(`[order submission] Order`) }</button>
```

### Body

The text body must always be written in [CLDR plural rule `other`](https://cldr.unicode.org/index/cldr-spec/plural-rules). This allows a TIK to avoid branched statements like ICU plural arguments.

Outside of [placeholders](#placeholders), the body is free-form text with no further syntactic restrictions. Curly braces `{` and `}` are reserved for placeholder syntax. To include a literal `{`, `}`, or `\` in the body, it must be escaped with a preceding reverse-solidus `\`:

- `\{` = literal `{`
- `\}` = literal `}`
- `\\` = literal `\`
- `\\\{\}` = literal `\{}`

Square brackets `[` and `]` may appear freely in the body (they are only special at the start of a TIK).

### Placeholders

Placeholders keep TIKs readable while remaining auto-translatable to the ICU message format. The following TIK uses multiple placeholders for different data types:

```
Today {name} earned {currency} for completing {# tasks} in section '{text}' at {time-short}.
```

- `{text}` [Text placeholder](#string-placeholders)
- `{name}` [Text placeholder with gender information](#string-placeholders-with-gender)
- `{integer}` Integer
- `{number}` Number
- `{# ...}` [Cardinal pluralization](#cardinal-pluralization)
- `{ordinal}` Ordinal pluralization
- `{date-full}` Date placeholder
- `{date-long}` Date placeholder
- `{date-medium}` Date placeholder
- `{date-short}` Date placeholder
- `{time-full}` Time placeholder
- `{time-long}` Time placeholder
- `{time-medium}` Time placeholder
- `{time-short}` Time placeholder
- `{currency}` Currency

### Cardinal Pluralization

A pluralization statement begins with `{#` and ends with `}`. The `#` serves as the placeholder where the numeric value is rendered in the generated ICU message. Everything between `#` and the closing `}` is the statement's content, which may be empty (`{#}`) or non-empty (`{# messages}`, `{#件のメッセージ}`). The content may include anything that is not explicitly forbidden (see [invariants](#cardinal-pluralization---syntactic-invariants)).

The contents may contain any number of placeholders:

```
You had {# messages marked as {text} at {time-long}}
```

```
You had {# tasks} assigned at {time-short}.
```

#### Cardinal Pluralization - Blank Pluralization Example

Languages without grammatical plural forms, such as Japanese (`ja`) or Chinese (`zh`), may use the blank `{#}` syntax, where only the numeric value is rendered:

TIK in source with native locale `ja`:
```
あなたには{#}件のメッセージがあります。
```

ICU generated by the TIK processor:
```
あなたには{var0, plural, other{#}}件のメッセージがあります。
```

Translated ICU for locale `en`:
```
You have {var0, plural, one{# message} other{# messages}}.
```

TIK in source with native locale `zh`:
```
{#}条新消息
```

ICU generated by the TIK processor:
```
{var0, plural, other{#}}条新消息
```

Translated ICU for locale `en`:
```
{var0, plural, one{# new message} other{# new messages}}
```

#### Cardinal Pluralization - Syntactic Invariants

1. Non-empty content must not consist solely of Unicode whitespace (as defined by [Unicode](https://unicode.org/charts/collation/chart_Whitespace.html)), and must not end with a Unicode whitespace character:

```
This TIK is illegal: {#  }
```

```
This TIK is illegal: {# messages }
```

2. Plural statements cannot be nested:

```
This TIK is illegal: {# first level {# second level}}
```

3. Content must not start with a placeholder:

```
This TIK is illegal: {#{integer}}
```

```
This TIK is illegal: {# {number}}
```

```
This TIK is illegal: {#{currency}}
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

An identifier that carries gender information (such as a person's name) is represented by [a string placeholder with gender](#string-placeholders-with-gender) rather than by a plain `{text}` placeholder, since gender affects grammar in gender-aware locales.

#### String Placeholders with Gender

String placeholders `{name}` carry gender information in addition to their string value. This placeholder represents arbitrary string values and is used for names and identifiers to enable correct translation in gender-aware locales.

```go
reader.String(
	`The journey began, {name} had embarked onto the ship.`, // TIK
	tokibundle.String{ Value: "John", Gender: tokibundle.GenderMale },
)
```

TIK does not define how gender information is attached to the placeholder; this is determined by the TIK processor.

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
| `{date-full}`   | `{var0, date, full}`                |
| `{date-long}`   | `{var0, date, long}`                |
| `{date-medium}` | `{var0, date, medium}`              |
| `{date-short}`  | `{var0, date, short}`               |
| `{time-full}`   | `{var0, time, full}`                |
| `{time-long}`   | `{var0, time, long}`                |
| `{time-medium}` | `{var0, time, medium}`              |
| `{time-short}`  | `{var0, time, short}`               |
| `{currency}`    | `{var0, number, ::currency/auto}`   |

The `...` stands for any content, meaning that the following TIK:

```
{# messages} in {# groups}
```

Encodes to the following ICU:

```
{var0, plural, other{# messages}} in {var0, plural, other{# groups}}
```

### Positional Argument Mapping

All placeholders are mapped positionally, meaning that the order of occurrence in the TIK is the order expected for argument inputs.

```
[report] By {time-short}, {name} received {# emails}.
```

All placeholders use the `var` prefix with a following positional index.

Generated ICU:
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
reader.String(`[report] By {time-short}, {name} received {# emails}.`,
	time.Now(), "Max", len(emailsReceived))
```

## Configuration Guidelines

The TIK specification defines guidelines only and imposes no strict format or requirements. The exact configuration format is left entirely to the processor implementation.

### Domains

A domain is the scope within which TIKs must be unique. The mechanism by which sources are assigned to a domain is left entirely to the TIK processor implementation: the specification does not mandate a configuration format, a discovery strategy, or a default boundary.
Implementations are free to treat the entire project as a single domain, derive domains from the directory layout, or accept explicit mappings.

Partitioning TIKs into multiple domains serves large-scale projects with many translations. One implementation defines domain scopes through a central configuration file:

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

Another implementation derives domains from marker files placed in the source tree. In such an implementation, a `.tikdomain` file marks a domain boundary covering its containing directory and all descendants, and nested `.tikdomain` files establish child domains whose descriptions append onto the parent's:

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

Beyond enforcing TIK uniqueness, a domain may carry a human-readable description that characterizes the scope, audience, or tone of its messages. Such descriptions may be forwarded by processors to LLM or human translators as additional context, improving translation accuracy beyond what the TIK alone conveys.

## Standards and Conventions

- Plural categories follow [Unicode CLDR](https://cldr.unicode.org/index/cldr-spec/plural-rules)
- Language codes follow [ISO 639-1](https://www.iso.org/iso-639-language-codes.html)
- Currency codes follow [ISO 4217](https://www.iso.org/iso-4217-currency-codes.html)
- Timestamps follow [RFC3339](https://www.rfc-editor.org/rfc/rfc3339.html)
- JSON examples follow [RFC8259](https://datatracker.ietf.org/doc/html/rfc8259)
- Date/Time formats follow [RFC1123](https://datatracker.ietf.org/doc/html/rfc1123)
