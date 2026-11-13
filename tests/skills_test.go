// Package tests checks the Rapibase Agent Skills before they ship.
//
// Two kinds of checks:
//
//   - Lint: skill directories flag skills that carry absolute URLs, download
//     commands or anything shaped like a secret. URLs come from the CLI at
//     runtime (`rapibase status --json`, `rapibase open`), so none of that may
//     appear in skills/ or README.md.
//   - Format: the Agent Skills spec (name, description, length, references one
//     level deep) plus this repo's own contract: the exact file list, a
//     metadata.version the CLI compares on `skill update`, and the Setup block
//     that must read the same in all four skills.
package tests

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const root = ".."

// expectedFiles is the skill → file list of the spec (§9.1). Adding a topic
// means adding its file here and linking it from the skill's SKILL.md.
var expectedFiles = map[string][]string{
	"rapibase-backend": {
		"SKILL.md",
		"references/auth.md",
		"references/config.md",
		"references/credentials.md",
		"references/design.md",
		"references/email.md",
		"references/pgvector.md",
		"references/push.md",
		"references/realtime.md",
		"references/rest.md",
		"references/rls.md",
		"references/rpc.md",
		"references/storage.md",
		"references/webhooks.md",
	},
	"rapibase-deploy": {
		"SKILL.md",
		"references/domains.md",
		"references/github-actions.md",
		"references/services.md",
		"references/site.md",
	},
	"rapibase-functions": {
		"SKILL.md",
		"references/runtime.md",
	},
	"rapibase-observe": {
		"SKILL.md",
		"references/errors.md",
		"references/restart.md",
	},
}

// ---------------------------------------------------------------- lint ---

type rule struct {
	name string
	re   *regexp.Regexp
}

var lintRules = []rule{
	{"absolute URL", regexp.MustCompile(`(?i)\bhttps?://`)},
	{"www host", regexp.MustCompile(`(?i)\bwww\.`)},
	// Domain-looking words. File names like package.json, config.toml,
	// schema.sql or orders.ts don't end in these suffixes, so they pass.
	{"domain name", regexp.MustCompile(`(?i)\b[a-z0-9][a-z0-9-]*\.(com|io|dev|app|net|org|ai|sh|co|cloud)\b`)},
	{"curl", regexp.MustCompile(`(?i)\bcurl\b`)},
	{"wget", regexp.MustCompile(`(?i)\bwget\b`)},
	{"pipe to a shell", regexp.MustCompile(`\|\s*(sh|bash|zsh)\b`)},
	{"base64 decode", regexp.MustCompile(`(?i)\bbase64\s+(-d|--decode)\b`)},
	{"PowerShell download (iwr)", regexp.MustCompile(`(?i)\biwr\s`)},
	{"PowerShell download (irm)", regexp.MustCompile(`(?i)\birm\s`)},
	{"Invoke-WebRequest/RestMethod", regexp.MustCompile(`(?i)\binvoke-(webrequest|restmethod)\b`)},
	{"JWT", regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}`)},
	{"Stripe-style key", regexp.MustCompile(`\b[sr]k_(live|test)`)},
	{"Rapibase token", regexp.MustCompile(`\brb(pat|dep|cli|mcp)_[A-Za-z0-9]{6,}`)},
	{"long hex string", regexp.MustCompile(`\b[0-9a-fA-F]{32,}\b`)},
}

// longToken finds base64-looking runs; a run is only a key if it mixes upper
// case, lower case and digits, which identifiers and API paths in the skills
// never do.
var longToken = regexp.MustCompile(`[A-Za-z0-9+/=_-]{40,}`)

func looksLikeBase64Secret(s string) bool {
	var upper, lower, digit bool
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			upper = true
		case r >= 'a' && r <= 'z':
			lower = true
		case r >= '0' && r <= '9':
			digit = true
		}
	}
	return upper && lower && digit
}

// lintedFiles: every file under skills/ plus README.md.
func lintedFiles(t *testing.T) []string {
	t.Helper()
	files := []string{filepath.Join(root, "README.md")}
	err := filepath.WalkDir(filepath.Join(root, "skills"), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking skills/: %v", err)
	}
	return files
}

func TestLint(t *testing.T) {
	for _, f := range lintedFiles(t) {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		sc := bufio.NewScanner(strings.NewReader(string(data)))
		sc.Buffer(make([]byte, 1024*1024), 1024*1024)
		n := 0
		for sc.Scan() {
			n++
			line := sc.Text()
			for _, r := range lintRules {
				if m := r.re.FindString(line); m != "" {
					t.Errorf("%s:%d: %s %q", rel(f), n, r.name, m)
				}
			}
			for _, m := range longToken.FindAllString(line, -1) {
				if looksLikeBase64Secret(m) {
					t.Errorf("%s:%d: long base64-like string %q", rel(f), n, m)
				}
			}
		}
	}
}

// lintLine reports the first rule a line breaks, or "".
func lintLine(line string) string {
	for _, r := range lintRules {
		if r.re.MatchString(line) {
			return r.name
		}
	}
	for _, m := range longToken.FindAllString(line, -1) {
		if looksLikeBase64Secret(m) {
			return "long base64-like string"
		}
	}
	return ""
}

// TestLintRules keeps the lint honest: it must flag each bad sample and
// leave the ordinary skill text alone. Key-like samples are built at runtime
// so this file holds nothing that looks like a key either.
func TestLintRules(t *testing.T) {
	bad := []string{
		"see " + "https" + "://x",
		"open " + "http" + "://localhost:3000",
		"go to " + "www" + ".example",
		"the site " + "example" + ".com",
		"docs at " + "foo" + ".dev",
		"cu" + "rl -s /x",
		"wg" + "et /x",
		"install | " + "sh",
		"install |" + "bash",
		"base64" + " -d file",
		"iw" + "r /x",
		"ir" + "m /x",
		"ey" + "J" + strings.Repeat("a", 12),
		"sk" + "_live",
		"rb" + "pat_" + "abcdef12",
		strings.Repeat("ab", 16),
		strings.Repeat("aB3", 14),
	}
	for _, s := range bad {
		if lintLine(s) == "" {
			t.Errorf("lint did not flag %q", s)
		}
	}
	good := []string{
		"package.json, config.toml, schema.sql, orders.ts, rapibase.json, bun.lock, Info.plist",
		"confirm the name first; firm commitments",
		"rapibase api post /api/v1/storage/upload/sign/my-bucket/avatar.png",
		"NEXT_PUBLIC_RAPIBASE_ANON_KEY and RAPIBASE_SERVICE_KEY",
		"a deploy token (it starts with rbdep_) and RAPIBASE_TOKEN",
		"uvicorn main:app --host 0.0.0.0 --port $PORT",
		"| `bash` | shell | shared |",
		"RAPIBASE_URL.replace(/^http/, 'ws')",
	}
	for _, s := range good {
		if r := lintLine(s); r != "" {
			t.Errorf("lint flagged ordinary text %q as %s", s, r)
		}
	}
}

// TestOnlyMarkdown: no executable scripts in v1 (the CLI does that work).
func TestOnlyMarkdown(t *testing.T) {
	err := filepath.WalkDir(filepath.Join(root, "skills"), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "scripts" {
				t.Errorf("%s: no scripts/ directories in v1", rel(p))
			}
			return nil
		}
		if filepath.Ext(p) != ".md" {
			t.Errorf("%s: only Markdown files are allowed under skills/", rel(p))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// -------------------------------------------------------------- format ---

var (
	nameRe    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	semverRe  = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
	linkRe    = regexp.MustCompile(`\]\(([^)\s]+)\)`)
	useWhenRe = regexp.MustCompile(`(?i)\buse (only )?when\b`)
)

type frontmatter struct {
	fields   map[string]string
	metadata map[string]string
}

// parseFrontmatter reads the small YAML subset the skills use: top-level
// "key: value" pairs and one nested "metadata:" map of strings.
func parseFrontmatter(content string) (*frontmatter, string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return nil, "", fmt.Errorf("no frontmatter: the file must start with ---")
	}
	end := strings.Index(content[4:], "\n---\n")
	if end < 0 {
		return nil, "", fmt.Errorf("frontmatter is not closed with ---")
	}
	head, body := content[4:4+end], content[4+end+5:]
	fm := &frontmatter{fields: map[string]string{}, metadata: map[string]string{}}
	inMeta := false
	for i, line := range strings.Split(head, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indented := strings.HasPrefix(line, " ")
		k, v, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			return nil, "", fmt.Errorf("frontmatter line %d is not key: value", i+1)
		}
		k, v = strings.TrimSpace(k), unquote(strings.TrimSpace(v))
		switch {
		case indented && inMeta:
			fm.metadata[k] = v
		case indented:
			return nil, "", fmt.Errorf("frontmatter line %d: unexpected indentation", i+1)
		default:
			inMeta = k == "metadata" && v == ""
			if !inMeta {
				if _, dup := fm.fields[k]; dup {
					return nil, "", fmt.Errorf("frontmatter key %q appears twice", k)
				}
				fm.fields[k] = v
			}
		}
	}
	return fm, body, nil
}

func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	return s
}

func skillDirs(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		t.Fatalf("reading skills/: %v", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		} else {
			t.Errorf("skills/%s: only skill directories belong in skills/", e.Name())
		}
	}
	return dirs
}

func TestExpectedFiles(t *testing.T) {
	var want []string
	for skill, files := range expectedFiles {
		for _, f := range files {
			want = append(want, skill+"/"+f)
		}
	}
	var got []string
	base := filepath.Join(root, "skills")
	err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			r, _ := filepath.Rel(base, p)
			got = append(got, filepath.ToSlash(r))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(want, "\n") != strings.Join(got, "\n") {
		t.Errorf("skills/ file list differs from the spec table\nwant:\n  %s\ngot:\n  %s",
			strings.Join(want, "\n  "), strings.Join(got, "\n  "))
	}
}

func TestSkillFormat(t *testing.T) {
	dirs := skillDirs(t)
	if len(dirs) != len(expectedFiles) {
		t.Errorf("found %d skills, want %d", len(dirs), len(expectedFiles))
	}
	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			if _, ok := expectedFiles[dir]; !ok {
				t.Errorf("unexpected skill directory %q", dir)
			}
			path := filepath.Join(root, "skills", dir, "SKILL.md")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("missing SKILL.md: %v", err)
			}
			content := string(data)
			fm, body, err := parseFrontmatter(content)
			if err != nil {
				t.Fatal(err)
			}

			name := fm.fields["name"]
			if name != dir {
				t.Errorf("name %q must equal the folder name %q", name, dir)
			}
			if len(name) == 0 || len(name) > 64 || !nameRe.MatchString(name) {
				t.Errorf("name %q: 1-64 chars, lowercase letters, digits and single hyphens", name)
			}

			desc := fm.fields["description"]
			if n := len([]rune(desc)); n < 1 || n > 1024 {
				t.Errorf("description has %d chars, want 1-1024", n)
			}
			if strings.ContainsAny(desc, "<>") {
				t.Errorf("description must not contain angle brackets")
			}
			if !useWhenRe.MatchString(desc) {
				t.Errorf("description must say when to use the skill (\"Use when ...\")")
			}

			if v := fm.metadata["version"]; !semverRe.MatchString(v) {
				t.Errorf("metadata.version %q is missing or not semver", v)
			}

			if lines := strings.Count(content, "\n") + 1; lines >= 500 {
				t.Errorf("SKILL.md has %d lines, want < 500", lines)
			}
			if words := len(strings.Fields(body)); words > 3500 {
				t.Errorf("SKILL.md body has %d words (~%d tokens); keep it under ~5000 tokens", words, words*4/3)
			}
		})
	}
}

// TestReferences: references/ is one level deep, every link to a local file
// resolves, and every reference is linked from its SKILL.md.
func TestReferences(t *testing.T) {
	for _, dir := range skillDirs(t) {
		t.Run(dir, func(t *testing.T) {
			skillRoot := filepath.Join(root, "skills", dir)
			refDir := filepath.Join(skillRoot, "references")
			entries, err := os.ReadDir(refDir)
			if err != nil {
				t.Fatalf("references/: %v", err)
			}
			skillMD, err := os.ReadFile(filepath.Join(skillRoot, "SKILL.md"))
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range entries {
				if e.IsDir() {
					t.Errorf("references/%s: references must be one level deep (no subfolders)", e.Name())
					continue
				}
				if !strings.Contains(string(skillMD), "](references/"+e.Name()+")") {
					t.Errorf("references/%s is not linked from SKILL.md", e.Name())
				}
			}

			_ = filepath.WalkDir(skillRoot, func(p string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				data, err := os.ReadFile(p)
				if err != nil {
					return err
				}
				for _, m := range linkRe.FindAllStringSubmatch(string(data), -1) {
					target := m[1]
					if strings.HasPrefix(target, "#") {
						continue
					}
					target, _, _ = strings.Cut(target, "#")
					if filepath.IsAbs(target) || strings.Contains(target, ":") {
						t.Errorf("%s: link %q must be a relative path", rel(p), m[1])
						continue
					}
					resolved := filepath.Join(filepath.Dir(p), filepath.FromSlash(target))
					if _, err := os.Stat(resolved); err != nil {
						t.Errorf("%s: link %q points to a missing file", rel(p), m[1])
					}
					if r, err := filepath.Rel(skillRoot, resolved); err != nil || strings.HasPrefix(r, "..") {
						t.Errorf("%s: link %q leaves the skill folder (skills are installed one by one)", rel(p), m[1])
					}
				}
				return nil
			})
		})
	}
}

// TestSetupBlockIsShared: the "## Setup" section is the common block (CLI
// first, MCP fallback, the folder's project, the fixed rules). It must read
// the same in every skill.
func TestSetupBlockIsShared(t *testing.T) {
	var first, firstDir string
	for _, dir := range skillDirs(t) {
		data, err := os.ReadFile(filepath.Join(root, "skills", dir, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		block := section(strings.ReplaceAll(string(data), "\r\n", "\n"), "## Setup")
		if block == "" {
			t.Errorf("%s: SKILL.md has no \"## Setup\" section", dir)
			continue
		}
		for _, must := range []string{"rapibase whoami", "npx rapibase@latest init --agent", "login_pending",
			".rapibase/project.json", "--project", "MCP", ".env", "explicit request", "git push"} {
			if !strings.Contains(block, must) {
				t.Errorf("%s: Setup block is missing %q", dir, must)
			}
		}
		if first == "" {
			first, firstDir = block, dir
			continue
		}
		if block != first {
			t.Errorf("%s: Setup block differs from %s's; it must be identical in all skills", dir, firstDir)
		}
	}
}

// section returns the text from a "## " heading to the next one.
func section(content, heading string) string {
	i := strings.Index(content, "\n"+heading)
	if i < 0 {
		return ""
	}
	rest := content[i+1:]
	if j := strings.Index(rest[len(heading):], "\n## "); j >= 0 {
		return rest[:len(heading)+j]
	}
	return rest
}

func rel(p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(r)
}
