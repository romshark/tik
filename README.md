# TIK - Textual Internationalization Key

"TIK" is an abbreviation for "Textual Internationalization Key".
A TIK is simultaneously the source of truth for translation and a unique message identifier within a domain.

TIKs make translation keys human-readable by closely reflecting the actual text shown to the end users in the source code. This improves context for translators, enables programmatic generation of
[ICU messages](https://unicode-org.github.io/icu/userguide/format_parse/messages/), and supports better automation and CI/CD integration.

TIK enables more efficient workflows by integrating TIK processors with CI and LLMs to give developers immediate feedback on i18n issues before they hit production. It reduces costs by minimizing reliance on human translators and eases pressure on them by offloading routine tasks, allowing experts to focus more on quality assurance.

![TIK i18n workflow](https://github.com/romshark/tik/blob/main/tik-i18n-workflow.svg)

TIK is designed to be agnostic to both programming languages and natural languages used in application source code.

TIP: Check out the official [TIK cheatsheet](https://romshark.github.io/tik-cheatsheet/).

**Table of Contents**

- [Specification](#specification)
- [Problem](#problem)
  - [Key-based Translation](#key-based-translation)
  - [ICU Messages](#icu-messages)
- [Limitations](#limitations)
- [FAQ](#faq)
- [Special Thanks](#special-thanks)

## Specification

The normative rules for TIK syntax, ICU encoding, and configuration guidelines are maintained in [SPECIFICATION.md](SPECIFICATION.md).

## Problem

Internationalization (i18n) and localization (l10n) are hard — and most developers avoid them. Supporting multiple languages and regions demands significant effort, expensive tooling, complex error-prone workflows with slow feedback loops, and discipline that many teams are unable to take on.

- Translators often work with vague context, leading to broken translations.
- Messages get over-abstracted for reuse breaking grammar and structure in many languages.
- Automation is limited by missing metadata and pipelines developers lack control over.
- The feedback loop is slow, brittle, and disconnected from CI/CD.

The result is missing or poor i18n and l10n that signals lack of polish, undermines
user trust, alienates global audiences and subsequently blocks adoption and growth.

### Key-based Translation

Traditional internationalization relies heavily on key-based systems, where developers assign abstract message identifiers (e.g. `"dashboard.report.summary"`) to translated
strings stored in external files.

```go
i18n.ByKey("dashboard.report.summary", numberOfMessages, dateTime)
```

Keys offer clear benefits, such as:

- **Separation of concerns -** Developers reference keys, while translators manage the actual text.
- **Reusability** - the same message can be used across different contexts or interfaces.
- **Dynamic updates** - translation changes go live immediately without redeployment.
- **Integration** - keys work seamlessly with most existing localization infrastructure.

However, key-based i18n introduces an abstraction layer between the source code and the actual text, making it harder for developers to immediately understand what message is being displayed - and in what form.

Naming is inherently hard - and coming up with meaningful, consistent translation keys can be difficult, especially at scale. Poorly chosen keys often lead to confusion, redundancy, or fragile reuse patterns.

TIKs, by contrast, embed the meaning directly in the code using a naturally readable and self-explanatory format that serves as source of truth for the i18n pipeline:

```go
reader.String(`You had {# messages} at {time-short}.`, numberOfMessages, dateTime)
```

### ICU Messages

[ICU messages](https://support.crowdin.com/icu-message-syntax) are a powerful internationalization tool but are too complex, unreadable and error-prone when used directly inside the application source code.

Consider the following example in Go:

```go
i18n.Text(`You had {numberOfMessages, plural,
	=0 {no messages}
	one {# message}
	other {# messages}
} at {time, date, jm}.`, numberOfMessages, dateTime)
```

With TIK, developers write simple, readable keys and still get the full power of ICU under the hood.

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

## FAQ

Frequently asked questions are maintained in [FAQ.md](FAQ.md).

## Special Thanks

Special thanks to Muthu Kumar ([@MKRhere](https://github.com/MKRhere))!
