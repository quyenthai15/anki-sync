# Card Generation Reference

Use the profile in `anki-sync profiles` as the source of truth for `deck`, `level`, `pronunciation`, `rules`, `fields`, and `audio`. The rules below apply on top of the profile.

## Field Rules

1. **Only profile fields.** Generate exactly the keys in `profile.fields`. Never carry over fields from another language (e.g. no `Kanji` for French).
2. **Sentence fields** follow the profile's language codes:
   - Target sentence: `Sentence_<TARGET_LANG>` (e.g. `Sentence_JP`, `Sentence_FR`), in the language of `target_lang`.
   - Translation sentence: `Sentence_<TRANSLATION_LANG>` (e.g. `Sentence_VN`, `Sentence_EN`), using the code from `translation_lang`.
3. **Audio** is defined by `profile.audio` as `"<AnkiAudioField>": "<SourceTextField>"`. The source field must be a generated, non-empty field. Do not invent audio fields.

## Pronunciation Formats

- **French:** IPA enclosed in slashes, e.g. `/livʁ/`, `/flɑ.ne/`.
- **Japanese:** Kana with Tokyo pitch accent in brackets, e.g. `だいがく [0]`, `たまご [2]`. Follow the mora count.
- **Mandarin (if configured):** Pinyin with tone marks.

If you are not confident in a reading or pitch accent, say so in the preview instead of guessing silently.

## Level Constraints

- Use the vocabulary and grammar allowed by `profile.level` (e.g. N5-N4 for Japanese, A0-A1 for French).
- Keep example sentences simple and follow `profile.rules`.

## Examples

### Japanese (`japanese` profile)

```yaml
- Expression: たまご
  Reading: たまご [2]
  Meaning: trứng
  Kanji: 卵
  Type: Noun
  Sentence_JP: あさごはんに たまごを たべます。
  Sentence_VN: Tôi ăn trứng vào bữa sáng.
```

### French (`french` profile)

```yaml
- Expression: le livre
  Reading: /livʁ/
  Meaning: cuốn sách
  Type: Nom (m)
  Sentence_FR: Je lis un livre intéressant ce soir.
  Sentence_EN: I am reading an interesting book tonight.
```

Note: the French example has no `Kanji` field, and the translation field is `Sentence_EN` because that profile uses `translation_lang: en`.
