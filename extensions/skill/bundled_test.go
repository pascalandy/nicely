package skill

import (
	"regexp"
	"testing"
)

// Released skills link to their own docs, while development builds use main.
func TestBundledDocumentationRevision(t *testing.T) {
	for _, tc := range []struct{ version, revision string }{
		{"", "main"},
		{"v0.0.0-dev", "main"},
		{"v0.27.0", "v0.27.0"},
		{"v0.27.0-rc.1", "v0.27.0-rc.1"},
		{"v0.26.1-0.20261103123456-0895b8b325fe", "main"},
	} {
		t.Run(tc.version, func(t *testing.T) {
			t.Setenv("NCLY_VERSION", tc.version)
			for _, skill := range bundledSkills() {
				if skill.Name != "nicely-extensions" {
					continue
				}
				links := regexp.MustCompile(`https://github\.com/pascalandy/nicely/blob/([^/]+)/`).FindAllStringSubmatch(skill.content, -1)
				if len(links) == 0 {
					t.Fatal("the extensions skill has no documentation links")
				}
				for _, link := range links {
					if link[1] != tc.revision {
						t.Errorf("link %s uses revision %q, want %q", link[0], link[1], tc.revision)
					}
				}
			}
		})
	}
}
