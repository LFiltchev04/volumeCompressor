package CephStorage

import (
	"path"
	"testing"

	"github.com/opencontainers/go-digest"
)

func TestPathForManifestTagLinks(t *testing.T) {
	revision, err := digest.Parse("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		spec     pathSpec
		want     string
		wantLink string
	}{
		{
			name: "tag index entry directory",
			spec: manifestTagIndexEntryPathSpec{
				name:     "tst",
				tag:      "v0",
				revision: revision,
			},
			want:     "/docker/registry/v2/repositories/tst/_manifests/tags/v0/index/sha256/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantLink: "/docker/registry/v2/repositories/tst/_manifests/tags/v0/index/sha256/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855/link",
		},
		{
			name: "tag current link",
			spec: manifestTagCurrentPathSpec{
				name: "tst",
				tag:  "v0",
			},
			want: "/docker/registry/v2/repositories/tst/_manifests/tags/v0/current/link",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PathFor(tt.spec)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("PathFor() = %q, want %q", got, tt.want)
			}
			if tt.wantLink != "" {
				if gotLink := path.Join(got, "link"); gotLink != tt.wantLink {
					t.Fatalf("index link path = %q, want %q", gotLink, tt.wantLink)
				}
			}
		})
	}
}
