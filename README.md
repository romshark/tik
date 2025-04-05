**Author:** Roman Scharkov <roman.scharkov@gmail.com>;
**Version:** 0.2;
**Last updated:** 2025-04-05;

# TIK - Textual Internationalization Key

**Table of Contents**

- [Introduction](#introduction)
- [Problem](#problem)
  - [Key-based Translation](#key-based-translation)
  - [ICU Messages](#icu-messages)
- [TIK Syntax Rules](#tik-syntax-rules)
  - [Magic Constants](#magic-constants)
  - [String Placeholders](#string-placeholders)
    - [String Placeholders with Gender and Pluralization](#string-placeholders-with-gender-and-pluralization)
    - [String Placeholder Invariants](#string-placeholder-invariants)
- [ICU Encoding](#icu-encoding)
- [ICU Encoding – Positional Mapping](#icu-encoding--positional-mapping)
  - [ICU Encoding - String Placeholders](#icu-encoding---string-placeholders)
  - [ICU Encoding - Gender Agreement](#icu-encoding---gender-agreement)
  - [ICU Encoding - Cardinal Pluralization](#icu-encoding---cardinal-pluralization)
  - [ICU Encoding - Ordinal Pluralization](#icu-encoding---ordinal-pluralization)
  - [ICU Encoding - Time Placeholders](#icu-encoding---time-placeholders)
  - [ICU Encoding - Currency](#icu-encoding---currency)
  - [ICU Encoding - Number](#icu-encoding---number)
- [Configuration Guidelines](#configuration-guidelines)
  - [Magic Constant Customization](#magic-constant-customization)
  - [Domains](#domains)
- [Limitations](#limitations)
  - [Limitations of Algorithmic ICU Message Generation](#limitations-of-algorithmic-icu-message-generation)
- [Standards and Conventions](#standards-and-conventions)
- [FAQ](#faq)
  - [Is this overcomplication really worth it and aren't simple keys enough?](#is-this-overcomplication-really-worth-it-and-arent-simple-keys-enough)
  - [How about just preloading translation texts by key using IDE plugins?](#how-about-just-preloading-translation-texts-by-key-using-ide-plugins)
  - [Could Fluent be used instead of ICU?](#could-fluent-be-used-instead-of-icu)
  - [Why use masculine gender by default instead of the neutral `they`?](#why-use-masculine-gender-by-default-instead-of-the-neutral-they)

## Introduction

"TIK" is an abbreviation for "Textual Internationalization Key".
TIKs make translation keys in source code more readable, provide better context for
translators, and enable programmatic generation of
[ICU messages](https://unicode-org.github.io/icu/userguide/format_parse/messages/).

## Problem

Internationalization is hard - and most developers avoid it.
Supporting multiple languages and regions often requires significant effort,
costly tooling, complex workflows and discipline,
which many teams are unwilling or unable to take on.

Translators often lack context, leading to errors. Messages are fragmented for reuse,
which breaks grammar in many languages, etc.

All this often conveys lack of professionalism to the end user and reduces their trust
in the product.

### Key-based Translation

Traditional internationalization relies heavily on key-based systems, where developers
assign abstract identifiers (e.g. `"dashboard.report.summary"`) to strings
stored in external files.

```go
localize.Text("dashboard.report.summary", numberOfMessages, dateTime)
```

Keys offer clear benefits, such as:

- **Separation of concerns -** Developers reference keys,
  while translators manage the actual text.
- **Reusability** - the same message can be used across different contexts or interfaces.
- **Integration** - keys work seamlessly with most existing localization infrastructure.

However, key-based i18n introduces an abstraction layer between the source code
and the actual text, making it harder for developers to immediately understand what
message is being displayed - and in what form.

Naming is inherently hard - and coming up with meaningful, consistent translation keys
can be difficult, especially at scale. Poorly chosen keys often lead to confusion,
redundancy, or fragile reuse patterns.

TIKs, by contrast, embed the meaning directly in the code using a naturally readable
and self-explanatory format:

```go
localize.Text(`You had {2} messages at {3:45PM}.`, numberOfMessages, dateTime)
```

### ICU Messages

[ICU messages](https://support.crowdin.com/icu-message-syntax) are a powerful
internationalization tool but are too complex, unreadable and error-prone when used
directly inside the application source code.

Consider the following example:

```go
localize.Text(`You had {numberOfMessages, plural,
    =0 {no messages}
    one {# message}
    other {# messages}
} at {time, date, jm}.`, numberOfMessages, dateTime)
```

With TIK, developers write simple, readable keys and still get the full power of
ICU under the hood.

## TIK Syntax Rules

A TIK must always be written in
[CLDR plural rule `other`](https://cldr.unicode.org/index/cldr-spec/plural-rules)
and masculine gender. This allows a TIK to avoid conditional ICU select statements.

### Magic Constants

Magic constants allow TIKs to be easily readable yet auto-translatable to ICU.
Below is an example TIK that uses multiple magic constants.

```
Today {he} earned {$1.20} for completing {2} tasks in section '{"job"}' at {3:45PM}.
```

- Cardinal Pluralization (see [cardinal pluralization](#icu-encoding---cardinal-pluralization)):
  - `{2}`: cardinal plural
- Ordinal Pluralization (see [ordinal pluralization](#icu-encoding---ordinal-pluralization)):
  - `{4th}`: ordinal plural
- Gender (see [gender agreement](#icu-encoding---gender-agreement))
  - `{he}`: with variants:
    - `{his}`: possessive pronoun
    - `{him}`: object pronoun
    - `{himself}`: reflexive pronoun
- Time (see [time placeholders](#icu-encoding---time-placeholders)):
  - `{3:45PM}`
  - `{3:45:30PM}`
  - `{April 2}`
  - `{Apr 2}`
  - `{Apr 2025}`
  - `{Monday}`
  - `{April 2, 3:45PM}`
  - `{2025}`
  - `{April 2, 3:45:30PM}`
- Currency (see [currency](#icu-encoding---currency))
  - `{$1}`
  - `{$1.20}`
  - `{USD 1}`
  - `{USD 1.20}`
- Number: (see [number](#icu-encoding---number))
  - `{3}`

The constants (and any of their variants) are case insensitive.

This concept is inspired by the
[time formatting](https://cs.opensource.google/go/go/+/master:src/time/format.go;l=109)
constants used in Go’s standard library `time` package.

### String Placeholders

String placeholders `{"..."}` accept arbitrary string values within the quotes:

```
This can be {"anything"} or {"anyone"}, {"cheers"}.
```

The quoted text is not literal output! It serves as a label or hint about the kind of
content that might appear there (e.g., a person's name, an object, etc.):

TIK:

```
And so the journey began, {"John"} had embarked onto the ship.
```

generated ICU:

```
And so the journey began, {userName} had embarked onto the ship.
```

#### String Placeholders with Gender and Pluralization

A string placeholder may be infused with gender and pluralization information,
which isn't specified in the TIK but can be provided programmatically in the source code
as in the following example in Go:

```go
localize.Text(
    `And so the journey began, {"John"} had embarked onto the ship.`, // TIK
    localize.TextValue{ Value: "Ada", Gender: localize.Female },
)
```

TIK doesn't define how gender or plural information is attached to placeholders.
This is determined by the TIK processor, which inspects the provided values in the
source code and applies grammar rules as needed.

generated ICU:

```
And so the journey began, {userName_gender, select,
  other { {userName} had embarked onto the ship.}
}
```

ℹ️ Gender may affect translation in some languages:

| Language  | masculine         | feminine            |
| :-------- | :---------------- | :------------------ |
| Ukrainian | `John готовий`    | `Martha готова`     |
| Italian   | `John è pronto`   | `Martha è pronta`   |
| French    | `John est prêt`   | `Martha est prête`  |
| Spanish   | `John está listo` | `Martha está lista` |
| Russian   | `John готов`      | `Martha готова`     |

The translated ICU message for locale `uk` would be:

```
І так розпочалася подорож, {userName_gender, select,
  female { {userName} вирушила на корабель. }
  male { {userName} вирушив на корабель. }
  other { {userName} вирушило на корабель. }
}
```

#### String Placeholder Invariants

The string placeholder text body (i.e., the text between the braces) must not be empty.

```
This is an invalid TIK: {""}.
```

The string placeholder text body must not start or end with a Unicode whitespace character
(as defined by [Unicode](https://unicode.org/charts/collation/chart_Whitespace.html)):

```
This is an invalid TIK: {" foo "}.
```

The text body of a string placeholder must not contain any of: `\`, `{`, `}` and `"`.

```
This is an invalid TIK: {"abc\"def\"ghi"}.
```

```
This is an invalid TIK: {"abc{def}ghi"}.
```

## ICU Encoding

## ICU Encoding – Positional Mapping

All placeholders are mapped positionally, meaning that the order of occurrence in the TIK
is the order expected for parameter inputs.

```
By {3:45PM}, {"John"} received {2} emails.
```

Example in Go:

```go
localize.Text(`By {3:45PM}, {"John"} received {2} emails.`,
    time.Now(), "Max", len(emailsReceived))
```

### ICU Encoding - String Placeholders

A string placeholder with gender or plural behavior wraps all tokens to the right,
up to the first hard sentence boundary (`.`, `!`, `?`).:

### ICU Encoding - Gender Agreement

Constants such as `{he}` (and all of its variations), as well as any `{"..."}` strings
with gender information included affect the next word to the right and
include it in the ICU block:

```
{He} is awesome
```

```
{gender, select,
    other { {gender} is awesome}
}
```

### ICU Encoding - Cardinal Pluralization

A plural constant `{2}` wraps the placeholder and all tokens to
the right up to the first hard sentence boundary (`.`, `!`, `?`).:

```
{2} messages are read. {2} are pending.
```

```
{numMessages, plural,
  other {# messages are read.}
}
{numMessages, plural,
  other {# are pending.}
}
```

Expected information type includes both integers and floating point numbers
(e.g. in Go `int`, `float64`, etc.).

### ICU Encoding - Ordinal Pluralization

The constant `{4th}` represents an ordinal plural number.

| Value       | en-US   | de-DE  | uk-UA    |
| :---------- | :------ | :----- | :------- |
| `int(1)`    | 1st     | 1.     | 1-ше     |
| `int(2)`    | 2nd     | 2.     | 2-ге     |
| `int(3)`    | 3rd     | 3.     | 3-тє     |
| `int(4)`    | 4th     | 4.     | 4-те     |
| `int(5)`    | 5th     | 5.     | 5-те     |
| `int(7)`    | 7th     | 7.     | 7-ме     |
| `int(8)`    | 8th     | 8.     | 8-ме     |
| `int(101)`  | 101st   | 101.   | 101-ше   |
| `int(102)`  | 102nd   | 102.   | 102-ге   |
| `int(103)`  | 103rd   | 103.   | 103-тє   |
| `int(104)`  | 104th   | 104.   | 104-те   |
| `int(1000)` | 1,000th | 1.000. | 1 000-не |

Expected information type includes both integers and floating point numbers
(e.g. in Go `int`, `float64`, etc.).

The constant `{4th}` also accepts numbers combined with gender.

| Value       | `uk-UA (f)` | `uk-UA (m)` | `uk-UA (n)` | `de-DE (f/m/n)` |
| :---------- | :---------- | :---------- | :---------- | :-------------- |
| `int(1)`    | `1-ша`      | `1-ший`     | `1-ше`      | `1.`            |
| `int(2)`    | `2-га`      | `2-ий`      | `2-ге`      | `2.`            |
| `int(3)`    | `3-тя`      | `3-ій`      | `3-тє`      | `3.`            |
| `int(4)`    | `4-та`      | `4-ий`      | `4-те`      | `4.`            |
| `int(5)`    | `5-та`      | `5-ий`      | `5-те`      | `5.`            |
| `int(101)`  | `101-ша`    | `101-ший`   | `101-ше`    | `101.`          |
| `int(102)`  | `102-га`    | `102-ий`    | `102-ге`    | `102.`          |
| `int(103)`  | `103-тя`    | `103-ій`    | `103-тє`    | `103.`          |
| `int(104)`  | `104-та`    | `104-ий`    | `104-те`    | `104.`          |
| `int(1000)` | `1 000-та`  | `1 000-ий`  | `1 000-не`  | `1,000.`        |

### ICU Encoding - Time Placeholders

Time placeholders are automatically localized to the appropriate format
for the given locale and expect both date and time information (e.g. in Go `time.Time`).

In the examples below, the time `"2025-07-14T19:44:11Z"` is the value represented.

| Constant               | ICU        | en-US               | de-DE              | uk-UA              | Description       |
| :--------------------- | :--------- | :------------------ | :----------------- | :----------------- | :---------------- |
| `{3:45PM}`             | `jm`       | 7:44PM              | 19:44              | 19:44              | Short time        |
| `{3:45:30PM}`          | `jms`      | 7:44:11PM           | 19:44:11           | 19:44:11           | Time with seconds |
| `{April 2}`            | `MMMMd`    | July 15             | 15. Juli           | 15 липня           | Full month + day  |
| `{Apr 2}`              | `MMMd`     | Jul 15              | 15. Juli           | 15 лип.            | Abbr. month + day |
| `{Apr 2025}`           | `MMMy`     | Jul 2025            | Jul. 2025          | лип. 2025          | Full month + year |
| `{Monday}`             | `EEEE`     | Tuesday             | Dienstag           | Вiвторок           | Weekday only      |
| `{April 2, 3:45PM}`    | `MMMMdjm`  | July 15, 7:44 PM    | 15. Juli, 19:44    | 15 липня, 19:44    | Date + short time |
| `{2025}`               | `y`        | 2025                | 2025               | 2025               | Year only         |
| `{April 2, 3:45:30PM}` | `MMMMdjms` | July 15, 7:44:11 PM | 15. Juli, 19:44:11 | 15 липня, 19:44:11 | Full datetime     |

### ICU Encoding - Currency

Currency placeholders are automatically localized to the appropriate format
for the given locale and expect both amount and currency information
(e.g. in Go `currency.Amount`).

In the examples below, `$39,250.45 USD` (`en-US`) is the value represented.

| Constant     | en-US         | de-DE         | uk-UA              | Description       |
| :----------- | :------------ | :------------ | :----------------- | :---------------- |
| `{$1}`       | 39.250$       | 39.250 $      | 39 250 дол. США    | Rounded           |
| `{$1.20}`    | $39,250.45    | 39.250,45 $   | 39 250,45 дол. США | Full              |
| `{USD 1}`    | USD 39,250    | 39.250 USD    | 39 250 USD         | Rounded with code |
| `{USD 1.20}` | USD 39,250.45 | 39.250,45 USD | 39 250,45 USD      | Full with code    |

### ICU Encoding - Number

The number constant `{3}` localizes the integer value to the appropriate format
for the given locale:

| Value        | en-US     | de-DE     | uk-UA     |
| :----------- | :-------- | :-------- | --------- |
| int(1)       | 1         | 1         | 1         |
| int(2)       | 2         | 2         | 2         |
| int(1000)    | 1,000     | 1.000     | 1 000     |
| int(1234567) | 1,234,567 | 1.234.567 | 1 234 567 |

## Configuration Guidelines

The TIK specification defines guidelines only
and imposes no strict format or requirements.
The exact configuration format is left entirely to the processor implementation.

### Magic Constant Customization

Not all codebases are written in English. In some cases, developers may prefer to write
source code and comments in their native language. In such scenarios, the default
[TIK magic constants](#magic-constants), which are English-based,
may reduce the overall readability and coherence of the source text.

This is an example in German:

```
{He} hat das Paket um {3:45PM} bekommen.
Heute ist das die {4th}-schnellste Lieferung.
Die Kosten betragen {$1.20}
```

Naturally, this code would benefit from overwriting the default magic constants:

```toml
# localization.toml

[magic constants]
"he/his/him/himself" = "er/ihn/ihm"
"{3:45PM}" = "{15:45 Uhr}"
"4th" = "4./4te/4ter"
"$1.20" = "1,20€"
"$1" = "1€"
```

### Domains

In large-scale projects with lots of translations it might make sense to group
extracted texts into domains by defining the scopes of the domains in the configuration:

```toml
# localization.toml

[domains]
"domain_A" = [
  "/domain_a/...",
  "/templates/domain_a/_",
]

"domain_B" = [
  "/domain_b/...",
  "/templates/domain_b/_",
  "/models/domain_b/_",
]
```

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


### Limitations of Algorithmic ICU Message Generation

TIK processors avoid complex text analysis (NLP) and rely on simple,
rule-based logic to decide which text belongs inside ICU `select` or `plural` blocks
for gender and cardinal forms.

Semantic adjustments - like restructuring sentences to reflect plurality or pronoun
agreement - are deferred to more advanced systems, or in the worst case,
handled manually by human experts.

For example, consider the following TIK:

```
{2} files were deleted permanently.
```

A TIK processor would typically generate the following ICU message from the TIK above:

```
{ numFiles, plural, other { # files } } were deleted permanently.
```

This is structurally correct, but semantically broken: only “files” is conditionally
pluralized, while “were deleted permanently” remains outside the block regardless
of number.

The correct ICU message should include **all text affected by the plural logic**:

```
{ numFiles, plural, other { # files were deleted permanently. } }
```

Automatically generating this structure isn't algorithmically feasible
without full sentence understanding. For this reason, this responsibility is deferred
to the translation layer (e.g. LLM-based translation), which can jointly translate
and rewrite the ICU message with proper semantics:

```
{ numFiles, plural,
  =0 { No files were deleted permanently. }
  one { # file was deleted permanently. }
  other { # files were deleted permanently. }
}
```

## Standards and Conventions

- Plural categories follow [Unicode CLDR](https://cldr.unicode.org/index/cldr-spec/plural-rules)
- Language codes follow [ISO 639-1](https://www.iso.org/iso-639-language-codes.html)
- Currency codes follow [ISO 4217](https://www.iso.org/iso-4217-currency-codes.html)
- Timestamps follow [RFC3339](https://www.rfc-editor.org/rfc/rfc3339.html)

## FAQ

### Is this overcomplication really worth it and aren't simple keys enough?

The answer depends on perspective. While simple keys like `dashboard.newsfeed.summary`
offer clear benefits, such as:

- separation of concerns
- easy reuse across contexts
- and compatibility with existing localization systems

they also come with certain [limitations](#problem).
It is likely that, for the foreseeable future, code will continue to be written and
maintained primarily by humans. At the same time, large language models are demonstrating
increasing proficiency in translation tasks. The concept behind TIK is to define clear,
human-readable messages directly in the source code, delegating the complexity of
generating accurate ICU messages for various languages to language models.

To give you some context, only the last sentence of this answer was actually written
by a human.

### How about just preloading translation texts by key using IDE plugins?

While theoretically viable, this approach is inherently limited to IDEs that support
such a feature. Additionally, those IDEs/extensions must be compatible with your
specific translation file format and message encoding (e.g., ICU, Fluent, ARB).
It also breaks down entirely when browsing code outside the IDE - for example,
on GitHub - where no plugin can preload or resolve translation keys.

### Could Fluent be used instead of ICU?

[Fluent](https://projectfluent.org/) can be considered a worthy
[counterpart to the ICU MessageFormat](https://github.com/projectfluent/fluent/wiki/Fluent-and-ICU-MessageFormat)
and nothing speaks against using it as an alternative TIK backend.

### Why use masculine gender by default instead of the neutral `they`?

Valid point! The simple truth is that `he` is shorter than `she` and `they`.
Luckily, this is [configurable](#configuration).
