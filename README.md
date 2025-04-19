**Author:** Roman Scharkov <roman.scharkov@gmail.com>;
**Version:** 0.4;
**Last updated:** 2025-04-19;

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
  - [Text](#text)
  - [Magic Constants](#magic-constants)
  - [Cardinal Pluralization](#cardinal-pluralization)
    - [Cardinal Pluralization Invariants](#cardinal-pluralization-invariants)
  - [String Placeholders](#string-placeholders)
    - [String Placeholders with Gender and Pluralization](#string-placeholders-with-gender-and-pluralization)
    - [String Placeholder Invariants](#string-placeholder-invariants)
- [ICU Encoding](#icu-encoding)
- [ICU Encoding – Positional Mapping](#icu-encoding--positional-mapping)
  - [ICU Encoding - String Placeholders](#icu-encoding---string-placeholders)
    - [ICU Encoding - String Placeholders With Gender](#icu-encoding---string-placeholders-with-gender)
    - [ICU Encoding - String Placeholders With Pluralization](#icu-encoding---string-placeholders-with-pluralization)
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
i18n.Text(`You had {2 messages} at {3:45PM}.`, numberOfMessages, dateTime)
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
[ignored spaces] [optional context [ignored spaces]] [text body] [ignored spaces]
```

A TIK consists of an optional context and the required text while the surrounding
unicode spaces are ignored. Both the context and text body must not be empty.

### Context

The TIK context is an optional namespace used to disambiguate message keys.
It is not part of the message’s text body and hence must not be included in the
If a TIK starts with an opening square bracket `[` then everything up to the next
closing square bracket `]` is treated as the context.

⚠️ Do not confuse context with the message description.

```go
// description.
i18n.Text(`[context] Text.`)
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

TIKs are unique message keys within a domain. However, the same original message text can
have different meanings depending on usage. In such cases, context must be added to
separate two TIKs with a similar text body.

Example: a 

```html
<body>
  <h1>{
    // "save" as in "save from danger".  <--- HERE
    i18n.Text(`Save your planet`)
  }</h1>
  <p>{ i18n.Text(`Your planet is in grave danger. Be the hero who saves it!`) }</p>
  <dialog>
    <p>You're about to exit the simulation.</p>
    <form method="dialog">
      <button>{
        // "save" as in "save to file".  <--- HERE
        i18n.Text(`Save your planet`)
      }</button>
      <button>{
        // Cancel exiting the simulation.
        i18n.Text(`Cancel`)
      }</button>
    </form>
  </dialog>
</body>
```

In the example above, the web page contains two TIKs that will result in 1 ICU message
being produced: `Save your planet`. In English, the meaning of the word "save" depends
on context, which allows this message to be reused across different contexts. But other
languages such as German might require two separate messages:

- `"Rette deinen Planeten"` (as in "save your planet from danger")
- `"Speichere deinen Planeten"` (as in "save your planet to file")

Since 1 TIK can never refer to 2 different messages a new TIK must be created yet
sometimes the original text must be preserved. In this case we can apply a context
to either (or both) messages to disambiguate them:

```
// "save" as in "save to file".
i18n.Text(`[button.save] Save your planet`)
```

The resulting TIK defines the context `"button.save"` and
the text body `"Save your planet"`.

### Text

The text part must always be written in
[CLDR plural rule `other`](https://cldr.unicode.org/index/cldr-spec/plural-rules)
and neutral gender. This allows a TIK to avoid conditional ICU select statements.

### Magic Constants

Magic constants allow TIKs to be easily readable yet auto-translatable to ICU.
Below is an example TIK that uses multiple magic constants.

```
Today {they} earned {$1.20} for completing {2 tasks} in section '{"job"}' at {3:45PM}.
```

- String Placeholders (see [string placeholders](#string-placeholders))
  - `{"..."}`
- Cardinal Pluralization (see [cardinal pluralization](#icu-encoding---cardinal-pluralization)):
  - `{2 ...}`: cardinal plural
- Ordinal Pluralization (see [ordinal pluralization](#icu-encoding---ordinal-pluralization)):
  - `{4th}`: ordinal plural
- Gender (see [gender agreement](#icu-encoding---gender-agreement))
  - `{they}` (subjective) with variants:
    - `{them}`: objective
    - `{their}`: possessive adjective
    - `{theirs}`: possessive pronoun
    - `{themself}`: reflexive
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

### Cardinal Pluralization

A pluralization statement `{2 ...}` begins with `{2 ` and ends with `}`.
The `2` is the placeholder for the actual number value.
The contents `...` may contain any text that is not explicitly forbidden
(see [invariants](#cardinal-pluralization-invariants)).

The contents may contain any number of string placeholders and magic constants:

```
You have {2 {"apples"} and {"bananas"}}.
```

```
You had {2 of {their} tasks assigned at {3:45PM}}
```

#### Cardinal Pluralization Invariants

1. Plural statements must not begin and end with a Unicode whitespace character.
(as defined by [Unicode](https://unicode.org/charts/collation/chart_Whitespace.html)):

```
This TIK is illegal: {2  <- two spaces here}
```

```
This TIK is illegal: {2 space here-> }
```

2. Plural statements cannot be nested:

```
This TIK is illegal: {2 first level {2 second level}}
```

3. Plural statement contents cannot start with a magic constant:

```
This TIK is illegal: {2 {3}}
```

```
This TIK is illegal: {2 {USD 1}}
```

```
This TIK is illegal: {2 {their}}
```

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
i18n.Text(
    `And so the journey began, {"John"} had embarked onto the ship.`, // TIK
    i18n.String{ Value: "Ada", Gender: i18n.GenderFemale },
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
  female { {_0} вирушила на корабель. }
  male { {_0} вирушив на корабель. }
  other { {_0} вирушило на корабель. }
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
[report] By {3:45PM}, {"John"} received {2 emails}.
```

generated ICU:
```
By { 0_, time, short }, { 1_gender, select,
  female { {1_value} received {2_, plural,
    one {# email}
    other {# emails}
  }. }
  male { {1_value} received {2_, plural,
    one {# email}
    other {# emails}
  }. }
  other { {1_value} received {2_, plural,
    one {# email}
    other {# emails}
  }. }
}
```

Usage example in Go:

```go
i18n.Text(`[report] By {3:45PM}, {"John"} received {2 emails}.`,
  time.Now(), "Max", len(emailsReceived))
```

### ICU Encoding - String Placeholders

A simple string placeholder without any additional information produces a simple
ICU placeholder:

TIK:

```
You are on page: {"Home"}!
```

generated ICU:

```
You are on page: {0_}!
```

Usage example in Go:

```go
i18n.Text(`You are on page: {"Home"}!`,
  "Home & Garden")
```

```go
i18n.Text(`You are on page: {"Home"}!`,
  i18n.String{Value: "Home & Garden"})
```

#### ICU Encoding - String Placeholders With Gender

A string placeholder with gender information produces an ICU `select` expression:

TIK:

```
{"John"} modified the file.
```

generated ICU:

```
{ 0_gender, select,
  other { {0_value} }
} modified the file.
```

Usage example in Go:

```go
i18n.Text(`{"John"} modified the file.`,
  i18n.String{Value: "Martha", Gender: i18n.GenderFemale})
```

#### ICU Encoding - String Placeholders With Pluralization

A string placeholder with pluralization information produce an ICU `select` expression:

TIK:

```
{"Students"} submitted the form.
```

generated ICU:

```
{ 0_number, plural,
  other { {0_value} }
} submitted the form.
```

Usage example in Go:

```go
i18n.Text(`{"students"} submitted the form.`,
  i18n.String{Value: "teachers", Number: len(teachersWhoSubmitted)})
```

Even though for English this example seems nonsensical, for translation into other
languages this information may often be neccessary.

The translated ICU message in Ukrainian would be:

```
{ 0_number, plural,
  one { {0_value} подав форму. }
  few { {0_value} подали форму. }
  many { {0_value} подали форму. }
  other { {0_value} подали форму. }
}
```

And as you can see, the plurality of the string value does affect the sentence structure.

### ICU Encoding - Gender Agreement

Magic constants such as `{They}` (and all of its variations)
produce an ICU `select` expression:

```
{They} built it {themself}
```

```
{_0, select,
  male {He}
  female {She}
  other {They}
} built it {_1, select,
  male {himself}
  female {herself}
  other {themself}
}.
```

Casing is preserved exactly as written in the TIK:

- `They` -> `He` (titled)
- `they` -> `he` (lower case)
- `THEY` -> `HE` (upper case)

### ICU Encoding - Cardinal Pluralization

The `{2 ...}` cardinal pluralization statement is encoded into an ICU `plural` expression.
The `2` is replaced with the `#` number placeholder and the contents `...` are wrapped
into the `other` rule

TIK:

```
{2 messages are unread.} {2 are pending.}
```

ICU:

```
{numUnread, plural,
  other {# messages are read.}
}
{numPending, plural,
  other {# are pending.}
}
```

TIK:

```
{2 slots} remaining.
```

ICU:

```
{numSlots, plural,
  other {# slots}
} remaining.
```

Expected information type includes both integers and floating point numbers
(e.g. in Go `int`, `float64`, etc.).

### ICU Encoding - Ordinal Pluralization

The magic constant `{4th}` encodes into an ordinal plural number.

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

The magic constant `{3}` localizes the integer value to the appropriate format
for the given locale:

| Value          | en-US     | de-DE     | uk-UA     |
| :------------- | :-------- | :-------- | --------- |
| `int(1)`       | 1         | 1         | 1         |
| `int(2)`       | 2         | 2         | 2         |
| `int(1000)`    | 1,000     | 1.000     | 1 000     |
| `int(1234567)` | 1,234,567 | 1.234.567 | 1 234 567 |

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
{They} hat das Paket um {3:45PM} bekommen.
Heute ist das die {4th}-schnellste Lieferung.
Die Kosten betragen {$1.20}.
```

Naturally, this code would benefit from overwriting the default magic constants:

```json
{
  "magic constants": {
    "they/them/their/theirs/themself": "er/ihn/sein/seiner/sich",
    "{3:45PM}": "{15:45 Uhr}",
    "4th": "4./4te/4ter",
    "$1.20": "1,20€",
    "$1": "1€"
  }
}
```

Finally, the german source code would look a lot more readable to german speakers:

```
{Er} hat das Paket um {15:45 Uhr} bekommen.
Heute ist das die {4.}-schnellste Lieferung.
Die Kosten betragen {1,20€}.
```

### Domains

In large-scale projects with lots of translations it might make sense to group
extracted texts into domains by defining the scopes of the domains in the configuration:

```json
{
  "domains": {
    "domain_A": [
      "/domain_a/...",
      "/templates/domain_a/_"
    ],
    "domain_B": [
      "/domain_b/...",
      "/templates/domain_b/_",
      "/models/domain_b/_"
    ]
  }
}
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

## Standards and Conventions

- Plural categories follow [Unicode CLDR](https://cldr.unicode.org/index/cldr-spec/plural-rules)
- Language codes follow [ISO 639-1](https://www.iso.org/iso-639-language-codes.html)
- Currency codes follow [ISO 4217](https://www.iso.org/iso-4217-currency-codes.html)
- Timestamps follow [RFC3339](https://www.rfc-editor.org/rfc/rfc3339.html)
- JSON examples follow [RFC8259](https://datatracker.ietf.org/doc/html/rfc8259)

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

While theoretically viable, this approach is inherently limited to IDEs that support
such a feature. Additionally, those IDEs/extensions must be compatible with your
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
