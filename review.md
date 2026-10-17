# SPECIFICATION.md Review

- **Revision:** [SPECIFICATION.md](SPECIFICATION.md) v0.9.0 at commit `524e6f4`, reviewed 2026-10-17. Line numbers (`Lnnn`) track the working copy and are updated as fixes change the spec.
- **Scope:** internal consistency, consistency with the reference implementation [tik-go](tik-go/) at the same commit, and the standards the spec cites or should cite.
- **Method:** tik-go behaviour was checked by running its parser and ICU translator on the inputs listed in the [appendix](#appendix-tik-go-results). Every external link in the spec was fetched.

## Findings

Rows are ordered from most to least severe. Status is one of `Open`, `Fixed` or `Won't fix`.

| # | Category | Location | Issue | Status |
| :-- | :-- | :-- | :-- | :-- |
| 1 | Wrong example | [L346](SPECIFICATION.md#L346) | `{# messages} in {# groups}` encodes the second plural as `var0`. It must be `var1`, and tik-go emits `var1`. | Fixed |
| 2 | Wrong example | [L361-365](SPECIFICATION.md#L361-L365), [L310-313](SPECIFICATION.md#L310-L313) | "Generated ICU" has three errors: (a) the name is rendered twice, once inside `other { {var1} }` and again in `} {var1} received`; (b) `one {# email}` cannot be generated, because a TIK encodes only the `other` form (see the ICU Encoding table and the blank pluralization example); (c) whitespace inside a branch is literal text in ICU MessageFormat, so `other { {var1} }` renders as " Max ". The `uk` example at L310-313 has the same whitespace defect: a double space after the comma and a trailing space. | Fixed |
| 3 | Wrong example | [L371-376](SPECIFICATION.md#L371-L376) | Passes a plain `"Max"` for `{name}`, which contradicts [L282-295](SPECIFICATION.md#L282-L295). This is the anti-pattern described in the README's "missing metadata" problem statement. | Fixed |
| 4 | Inconsistency | [L322](SPECIFICATION.md#L322), [L310](SPECIFICATION.md#L310), [L361](SPECIFICATION.md#L361), [L357](SPECIFICATION.md#L357) | `{name}` has three different encodings: (a) the table gives `{var0, select, other{...}}`, which selects on the name string itself; (b) the examples use a separate `varN_gender` argument; (c) tik-go emits a plain `{var0}`. The naming rule at L357 mentions neither the `_gender` suffix nor that indexes are zero-based. | Fixed |
| 5 | Inconsistency | [L35](SPECIFICATION.md#L35), [L188](SPECIFICATION.md#L188), [L236](SPECIFICATION.md#L236) | "Empty" vs. plural invariant 1. L35 defines whitespace-only content as *empty*, and L188 allows empty plural content, so together they allow `{#  }`. Invariant 1 (L236) forbids it, but under that definition its wording ("non-empty content must not consist solely of whitespace") rules nothing out. Suggested wording: "The content is either absent (`{#}`) or contains at least one non-whitespace character and does not end with whitespace." | Open |
| 6 | Inconsistency | [L252](SPECIFICATION.md#L252), [L259](SPECIFICATION.md#L259) | Invariant 3 forbids content that *starts* with a placeholder, but its example `{# {number}}` starts with a space. Add "ignoring leading whitespace", which is what tik-go does. | Open |
| 7 | Inconsistency | [L153](SPECIFICATION.md#L153) | Lone reverse solidus: the spec says a literal `\` *must* be escaped, but tik-go accepts `C:\path` and a trailing `\` as literals. State whether `\` followed by any other character is an error or a literal. | Open |
| 8 | Inconsistency | [L153](SPECIFICATION.md#L153) | Escapes inside plural content are not specified: L153 covers only text "outside of placeholders". In tik-go, `\{` works inside plural content but `\}` closes the plural, so `{# items \} more}` fails with "unexpected directive closure". | Open |
| 9 | Inconsistency | [L2-3](SPECIFICATION.md#L2-L3) | The header still reads 0.9.0 / 2026-04-05, unchanged since `720baaa`. Fifteen commits have changed the spec since then, including the breaking change `4ad3099` (2026-04-07). | Open |
| 10 | Inconsistency | [L151](SPECIFICATION.md#L151), [L164](SPECIFICATION.md#L164), [L170-171](SPECIFICATION.md#L170-L171), [L286](SPECIFICATION.md#L286) | Terminology: (a) the headings say "String placeholder", the list at L170-171 says "Text placeholder", and L286 calls `{name}` a plain "String placeholder"; (b) L151 says "plural **rule** `other`", but `other` is a plural *category*, as L457 says; (c) "auto-translatable" (L164) names the TIK-to-ICU conversion, while "Translated ICU" (L214, L229, L307) means human translation. Use "encode" or "generate" for the conversion, as L209 and the ICU Encoding section already do. | Open |
| 11 | Inconsistency | [L86-88](SPECIFICATION.md#L86-L88) | The section "Context Uniqueness" is about TIK uniqueness. Its clause "as long as every occurrence shares the exact same context and body" is circular, because occurrences that differ are different TIKs. | Open |
| 12 | Inconsistency | [L211](SPECIFICATION.md#L211), [L325](SPECIFICATION.md#L325), [L361](SPECIFICATION.md#L361) | The generated ICU is formatted inconsistently: `other{#}`, `other {#}` (tik-go) and `{ var0, time, short }`. Pick one style, or state that whitespace outside message text is insignificant. | Open |
| 13 | Inconsistency | [L45](SPECIFICATION.md#L45), [L110](SPECIFICATION.md#L110), [L291](SPECIFICATION.md#L291), [L374](SPECIFICATION.md#L374) | The example APIs vary: `reader.String`, `i18n.Text`, and `tokibundle.String`, which belongs to one implementation and is never introduced. Use one neutral name. | Open |
| 14 | Inconsistency | [L1](SPECIFICATION.md#L1) | The header spells the author's surname "Scharkov", but the commits use "Sharkov". Confirm which spelling the spec should carry. | Open |
| 15 | Standards | [L317](SPECIFICATION.md#L317), [L455-462](SPECIFICATION.md#L455-L462) | ICU MessageFormat is never cited, although the ICU Encoding section is normative. Cite its syntax, and state that the target is MessageFormat 1, since Unicode MessageFormat 2.0 is now standardized in UTS #35 Part 9. | Open |
| 16 | Standards | [L335](SPECIFICATION.md#L335), [L459](SPECIFICATION.md#L459) | `::currency/auto` is not valid ICU skeleton syntax, because the `currency` stem requires a three-letter ISO 4217 code. Either use `{var0, number, currency}` with an argument that carries both the amount and its currency, or document `auto` as an extension that the processor resolves. Either change would also give the currently unused ISO 4217 reference (L459) a purpose. | Open |
| 17 | Standards | [L462](SPECIFICATION.md#L462) | RFC 1123 is the wrong standard. It defines a fixed English mail/HTTP date format, whereas `{date-*}` and `{time-*}` use ICU/CLDR styles that depend on the locale. Cite UTS #35 Part 4 (Dates) instead. | Open |
| 18 | Standards | [L460](SPECIFICATION.md#L460) | RFC 3339 "Timestamps": the only timestamp in the spec is the header date. Clarify what the reference applies to, or drop it. | Open |
| 19 | Standards | [L458](SPECIFICATION.md#L458) | ISO 639-1 was withdrawn in November 2023 and replaced by ISO 639:2023. The spec also uses *locales* (`ja`, `zh`, `uk`, "native locale") rather than bare language codes. Cite BCP 47 / Unicode locale identifiers (UTS #35 Part 1) instead; CLDR plural rules are keyed by these too. | Open |
| 20 | Standards | [L236](SPECIFICATION.md#L236), [L35](SPECIFICATION.md#L35) | The whitespace link returns 404, and a collation chart was never the definition of whitespace anyway. Cite the Unicode `White_Space` property (UAX #44, `PropList.txt`), which is what Go's `unicode.IsSpace` implements, and cite it where the term first appears (L35). | Open |
| 21 | Standards | [L151](SPECIFICATION.md#L151), [L457](SPECIFICATION.md#L457) | The CLDR plural rules link points to an informal CLDR page. The normative source is UTS #35 Part 3, "Language Plural Rules", which covers both cardinal and ordinal rules. | Open |
| 22 | Standards | [L7](SPECIFICATION.md#L7) | There is no reference to BCP 14 (RFC 2119 / RFC 8174) for "must" and "may", and the spec doesn't mark which sections are informative (the examples and Configuration Guidelines). | Open |
| 23 | Standards | [L32](SPECIFICATION.md#L32) | There is no formal grammar: (a) L32 shows `[body]` as optional although the body is required; (b) `[ ]` is used both to mark optional parts and as the literal context delimiters; (c) the prose doesn't say what happens to whitespace after the first separator character (tik-go strips all of it). An ABNF grammar (RFC 5234) would fix this and also pin down escapes and plural content. | Open |
| 24 | Underspecified | [L153-158](SPECIFICATION.md#L153-L158), [L167](SPECIFICATION.md#L167), [L317-335](SPECIFICATION.md#L317-L335) | The spec doesn't say how literal text is escaped in the generated ICU: (a) apostrophes: the spec's own example at L167, `'{text}'`, becomes quoted literal text in ICU unless each `'` is doubled; tik-go doubles them, but the spec doesn't require it; (b) `\{` and `\}`: tik-go emits raw braces, so `\{not a placeholder\}` becomes the ICU argument `{not a placeholder}`, and `\\\{\}` produces the invalid `\{}`; (c) `#` inside plural content: in `{# tasks in #general}`, ICU replaces both `#` with the number, so the literal one must be quoted as `'#'`. | Open |
| 25 | Underspecified | [L326](SPECIFICATION.md#L326) | `{ordinal}` encodes as `#th`, which is English-only and contradicts the README's claim that TIK is agnostic to natural languages. tik-go already makes the suffix configurable (`Config.OrdinalPluralOtherSuffix`), but a suffix can't express languages that put the ordinal marker first (Chinese 第#). | Open |
| 26 | Underspecified | [L164-184](SPECIFICATION.md#L164-L184) | The spec doesn't define: (a) the difference between `{integer}` and `{number}`; (b) the argument types for `{ordinal}`, date/time and `{currency}`, or whether `{#}` accepts non-integers; (c) the gender select keys (`male`/`female`/`other`, used in the `uk` example); (d) how many positional arguments `{name}` consumes. | Open |
| 27 | Underspecified | [L153](SPECIFICATION.md#L153), [L164-184](SPECIFICATION.md#L164-L184) | tik-go rejects the following cases, but the spec never says they are errors: unknown placeholders, wrong case (`{Text}`), whitespace inside braces (`{ text }`), a stray `}`, and an unclosed `{`. | Open |
| 28 | Underspecified | [L86-88](SPECIFICATION.md#L86-L88) | TIK identity: uniqueness is enforced at build time, so the spec needs to define when two TIKs are equal. (a) Is `[ order submission ]` the same as `[order submission]`? (b) Is whitespace inside the body significant, for example in multi-line raw strings? (c) Is Unicode normalization (UAX #15) applied? | Open |
| 29 | Underspecified | [L39](SPECIFICATION.md#L39), [L453](SPECIFICATION.md#L453) | Catalog key: the context is stripped from the ICU output, so `[a] Order` and `[b] Order` both produce the message `Order`. State that the full TIK is the catalog key, and that processors should pass the context on to translators, as L453 already says for domain descriptions. | Open |
| 30a | Underspecified | [L160](SPECIFICATION.md#L160) | A body without a context can't start with `[`, and there is no `\[` escape. Document the workaround (`[ctx] [1] …`). | Open |
| 30b | Underspecified | [L29-35](SPECIFICATION.md#L29-L35) | State that a TIK is the runtime value of the host-language string. JS template literals, for example, turn `\{` into `{` before a processor sees it. | Open |
| 30c | Underspecified | [L41](SPECIFICATION.md#L41), [L204](SPECIFICATION.md#L204), [README.md L104](README.md#L104) | "TIK processor" is defined only in README.md (L104). "Native locale" (L204) and "message description" (L41) are never defined. | Open |
| 30d | Underspecified | [L381-385](SPECIFICATION.md#L381-L385) | "Domain" is normative, because the uniqueness rule depends on it, but it is defined only in the section that calls itself "guidelines only". | Open |
| 30e | Underspecified | [L405](SPECIFICATION.md#L405) | In the JSON example, the `_` in `"/templates/domain_a/_"` is never explained (possibly a typo for `/...`). The example also doesn't say which domain wins where `/...` overlaps `/domain_a/...`. | Open |
| 31 | Typo/style | [L39](SPECIFICATION.md#L39) | Curly apostrophe in `message’s`, the only one in the spec. | Open |
| 32 | Typo/style | [L188](SPECIFICATION.md#L188), [L190](SPECIFICATION.md#L190) | "content" vs. "contents". | Open |
| 33 | Typo/style | [L343](SPECIFICATION.md#L343) | "Encodes" should be "encodes", because the sentence starts at L337. | Open |
| 34 | Typo/style | [L193](SPECIFICATION.md#L193) | Missing final period (compare L197). | Open |
| 35 | Typo/style | [L290](SPECIFICATION.md#L290) | Comma splice. "boarded the ship" reads better than "embarked onto the ship". | Open |
| 36 | Typo/style | [L291](SPECIFICATION.md#L291) | Not formatted with gofmt (`tokibundle.String{Value: …}`). | Open |
| 37 | Typo/style | [L418](SPECIFICATION.md#L418) | "append onto the parent's" should be "are appended to the parent's". | Open |
| 38 | Typo/style | [L453](SPECIFICATION.md#L453) | "LLM or human translators" should be "LLM-based or human translators". | Open |
| 39 | Typo/style | [L92](SPECIFICATION.md#L92) | A hyphen is used as a dash. | Open |
| 40 | Typo/style | [L50](SPECIFICATION.md#L50), [L153](SPECIFICATION.md#L153) | "reverse-solidus" should be "reverse solidus" (U+005C). | Open |
| 41 | Typo/style | [L460-462](SPECIFICATION.md#L460-L462) | "RFC3339" should be "RFC 3339"; the same applies to 8259 and 1123. | Open |
| 42 | Outside spec | [tik-go/tik.go L244-265](tik-go/tik.go#L244-L265) | tik-go accepts a plural that is never closed if it contains a placeholder: `{# messages from {text}` produces the invalid `{var0, plural, other {# messages from {var1}`, because the end-of-input path never checks `inPluralDirective`. | Open |
| 43 | Outside spec | [README.md L95](README.md#L95) | `{time, date, jm}` is invalid ICU, because `j` is allowed only in skeletons. It needs `::jm`. | Open |
| 44 | Outside spec | [FAQ.md L13](FAQ.md#L13) | The "ICU MessageFormat" link points to a Fluent wiki page. | Open |
| 45 | Outside spec | [tik-go/icu.go L48-53](tik-go/icu.go#L48-L53) | tik-go encodes `{name}` as a plain `{varN}` instead of `{varN_gender, select, other{{varN}}}`, which the spec now requires (see #4). Aligning it changes tik-go's output, which is a breaking change for its users. | Open |

## Appendix: tik-go results

Each input was run through `tik.Parser.Parse` and `tik.ICUTranslator.TIK2ICU` with `tik.DefaultConfig`.

| TIK input | tik-go result |
| :-------- | :------------ |
| `{# messages} in {# groups}` | `{var0, plural, other {# messages}} in {var1, plural, other {# groups}}` |
| `[report] By {time-short}, {name} received {# emails}.` | `By {var0, time, short}, {var1} received {var2, plural, other {# emails}}.` |
| L167 example, `… in section '{text}' at {time-short}.` | `… in section ''{var3}'' at {var4, time, short}.` |
| `\{not a placeholder\}` | `{not a placeholder}` |
| `\\\{\}` | `\{}` |
| `{# tasks in channel #general}` | `{var0, plural, other {# tasks in channel #general}}` |
| `{# items \} more}` | error: unexpected directive closure |
| `{# messages from {text}` | accepted: `{var0, plural, other {# messages from {var1}` |
| `C:\path`, `ends with \` | accepted unchanged |
| `{ text }`, `{Text}` | error: unknown placeholder |
| `[ order submission ] Order` | context token `[ order submission ]`, with the inner spaces kept |
| `You're {ordinal}` | `You''re {var0, selectordinal, other {#th}}` |

## References

- [ICU MessageFormat](https://unicode-org.github.io/icu/userguide/format_parse/messages/)
- [ICU number skeletons](https://unicode-org.github.io/icu/userguide/format_parse/numbers/skeletons.html)
- [UTS #35 (LDML)](https://www.unicode.org/reports/tr35/), Part 1: Unicode locale identifiers
- [UTS #35 Part 3: Language Plural Rules](https://www.unicode.org/reports/tr35/tr35-numbers.html#Language_Plural_Rules)
- [UTS #35 Part 4: Dates](https://www.unicode.org/reports/tr35/tr35-dates.html)
- [UTS #35 Part 9: MessageFormat](https://www.unicode.org/reports/tr35/tr35-messageFormat.html)
- [UAX #44: Unicode Character Database](https://www.unicode.org/reports/tr44/) and [PropList.txt](https://www.unicode.org/Public/UCD/latest/ucd/PropList.txt) (`White_Space`)
- [UAX #15: Unicode Normalization Forms](https://www.unicode.org/reports/tr15/)
- [BCP 47: Tags for Identifying Languages](https://www.rfc-editor.org/info/bcp47)
- [BCP 14: Key words for requirement levels](https://www.rfc-editor.org/info/bcp14)
- [RFC 5234: ABNF](https://www.rfc-editor.org/info/rfc5234)
- [ISO 639:2023](https://www.iso.org/standard/74575.html)
