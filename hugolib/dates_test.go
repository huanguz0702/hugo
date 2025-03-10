// Copyright 2021 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugolib

import (
	"fmt"
	"strings"
	"testing"
)

func TestDateFormatMultilingual(t *testing.T) {
	b := newTestSitesBuilder(t)
	b.WithConfigFile("toml", `
baseURL = "https://example.org"

defaultContentLanguage = "en"
defaultContentLanguageInSubDir = true

[languages]
[languages.en]
weight=10
[languages.nn]
weight=20
	
`)

	pageWithDate := `---
title: Page
date: 2021-07-18
---	
`

	b.WithContent(
		"_index.en.md", pageWithDate,
		"_index.nn.md", pageWithDate,
	)

	b.WithTemplatesAdded("index.html", `
Date: {{ .Date | time.Format ":date_long" }}
	`)

	b.Build(BuildCfg{})

	b.AssertFileContent("public/en/index.html", `Date: July 18, 2021`)
	b.AssertFileContent("public/nn/index.html", `Date: 18. juli 2021`)

}

func TestTimeZones(t *testing.T) {
	b := newTestSitesBuilder(t)
	b.WithConfigFile("toml", `
baseURL = "https://example.org"

defaultContentLanguage = "en"
defaultContentLanguageInSubDir = true

[languages]
[languages.en]
timeZone="UTC"
weight=10
[languages.nn]
timeZone="America/Antigua"
weight=20
	
`)

	const (
		pageTemplYaml = `---
title: Page
date: %s
lastMod: %s
publishDate: %s
expiryDate: %s
---	
`

		pageTemplTOML = `+++
title="Page"
date=%s
lastMod=%s
publishDate=%s
expiryDate=%s
+++
`

		shortDateTempl = `%d-07-%d`
		longDateTempl  = `%d-07-%d 15:28:01`
	)

	createPageContent := func(pageTempl, dateTempl string, quoted bool) string {
		createDate := func(year, i int) string {
			d := fmt.Sprintf(dateTempl, year, i)
			if quoted {
				return fmt.Sprintf("%q", d)
			}

			return d
		}

		return fmt.Sprintf(
			pageTempl,
			createDate(2021, 10),
			createDate(2021, 11),
			createDate(2021, 12),
			createDate(2099, 13), // This test will fail in 2099 :-)
		)
	}

	b.WithContent(
		// YAML
		"short-date-yaml-unqouted.en.md", createPageContent(pageTemplYaml, shortDateTempl, false),
		"short-date-yaml-unqouted.nn.md", createPageContent(pageTemplYaml, shortDateTempl, false),
		"short-date-yaml-qouted.en.md", createPageContent(pageTemplYaml, shortDateTempl, true),
		"short-date-yaml-qouted.nn.md", createPageContent(pageTemplYaml, shortDateTempl, true),
		"long-date-yaml-unqouted.en.md", createPageContent(pageTemplYaml, longDateTempl, false),
		"long-date-yaml-unqouted.nn.md", createPageContent(pageTemplYaml, longDateTempl, false),
		// TOML
		"short-date-toml-unqouted.en.md", createPageContent(pageTemplTOML, shortDateTempl, false),
		"short-date-toml-unqouted.nn.md", createPageContent(pageTemplTOML, shortDateTempl, false),
		"short-date-toml-qouted.en.md", createPageContent(pageTemplTOML, shortDateTempl, true),
		"short-date-toml-qouted.nn.md", createPageContent(pageTemplTOML, shortDateTempl, true),
	)

	const datesTempl = `
Date: {{ .Date | safeHTML  }}
Lastmod: {{ .Lastmod | safeHTML  }}
PublishDate: {{ .PublishDate | safeHTML  }}
ExpiryDate: {{ .ExpiryDate | safeHTML  }}

	`

	b.WithTemplatesAdded(
		"_default/single.html", datesTempl,
	)

	b.Build(BuildCfg{})

	expectShortDateEn := `
Date: 2021-07-10 00:00:00 +0000 UTC
Lastmod: 2021-07-11 00:00:00 +0000 UTC
PublishDate: 2021-07-12 00:00:00 +0000 UTC
ExpiryDate: 2099-07-13 00:00:00 +0000 UTC`

	expectShortDateNn := strings.ReplaceAll(expectShortDateEn, "+0000 UTC", "-0400 AST")

	expectLongDateEn := `
Date: 2021-07-10 15:28:01 +0000 UTC
Lastmod: 2021-07-11 15:28:01 +0000 UTC
PublishDate: 2021-07-12 15:28:01 +0000 UTC
ExpiryDate: 2099-07-13 15:28:01 +0000 UTC`

	expectLongDateNn := strings.ReplaceAll(expectLongDateEn, "+0000 UTC", "-0400 AST")

	// TODO(bep) create a common proposal for go-yaml, go-toml
	// for a custom date parser hook to handle these time zones.
	// JSON is omitted from this test as JSON does no (to my knowledge)
	// have date literals.

	// YAML
	// Note: This is with go-yaml v2, I suspect v3 will fail with the unquouted values.
	b.AssertFileContent("public/en/short-date-yaml-unqouted/index.html", expectShortDateEn)
	b.AssertFileContent("public/nn/short-date-yaml-unqouted/index.html", expectShortDateNn)
	b.AssertFileContent("public/en/short-date-yaml-qouted/index.html", expectShortDateEn)
	b.AssertFileContent("public/nn/short-date-yaml-qouted/index.html", expectShortDateNn)

	b.AssertFileContent("public/en/long-date-yaml-unqouted/index.html", expectLongDateEn)
	b.AssertFileContent("public/nn/long-date-yaml-unqouted/index.html", expectLongDateNn)

	// TOML
	// These fails: TOML (Burnt Sushi) defaults to local timezone.
	// TODO(bep) check go-toml
	b.AssertFileContent("public/en/short-date-toml-unqouted/index.html", expectShortDateEn)
	b.AssertFileContent("public/nn/short-date-toml-unqouted/index.html", expectShortDateNn)
	b.AssertFileContent("public/en/short-date-toml-qouted/index.html", expectShortDateEn)
	b.AssertFileContent("public/nn/short-date-toml-qouted/index.html", expectShortDateNn)

}
