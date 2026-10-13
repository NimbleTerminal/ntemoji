# Pinned Unicode CLDR data

These unmodified English annotation files come from **CLDR 48.2**, tag
`release-48-2` in Unicode's official repository:

| Local file | Upstream file | SHA-256 |
| --- | --- | --- |
| `en.xml` | [common/annotations/en.xml](https://github.com/unicode-org/cldr/blob/release-48-2/common/annotations/en.xml) | `8511aadd046fdba2f0ffe590266ced8bbf48175ad139b2675d85d7141057b235` |
| `en-derived.xml` | [common/annotationsDerived/en.xml](https://github.com/unicode-org/cldr/blob/release-48-2/common/annotationsDerived/en.xml) | `d76bd041c8c9e7b00b716aff8b7d9dbf509877010e191d5efd068be2553e066e` |

The snapshot includes emoji, sequences, and other annotated symbols. `type="tts"`
provides descriptive names; other annotations provide keywords separated by `|`.
CLDR omits U+FE0F emoji presentation selectors from its keys. Runtime lookups do
the same without changing the input returned for an unknown value.

From the repository root, regenerate the checked-in Go tables with:

```sh
task generate-cldr
```

The task runs `go generate ./...`, which can also be invoked directly.
Generation verifies the source checksums and uses no network. Ordinary builds
use `metadata_generated.go`; they do not read XML or run the generator. The test
suite checks that generated output exactly matches the pinned snapshot.

To update, select a published CLDR release, download the two files and its
`LICENSE`, update the release and hashes in `internal/cmd/gencldr/main.go` and this
document, regenerate, and run `task ci`. Review changes to names and keywords.
The generator currently accepts English only; adding other locales requires
explicit locale handling and fallback rules.

The data and generated tables are covered by [Unicode License v3](../../third_party/unicode/LICENSE.txt).
Keep the [attribution notice](../../third_party/unicode/NOTICE.txt) with redistributed data.
