package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

// make retrieve buildpackTomlPath="$PWD/../buildpack.toml" output="/tmp/metadata.json"

type GithubTagReponse struct {
	Name string `json:"name"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("missing inputs")
		os.Exit(0)
	}

	buildpackTomlPath := os.Args[1]
	output := os.Args[2]

	fmt.Printf("Nginx retrieve:\n\tbuildpackTomlPath: %s\n\toutput: %s\n", buildpackTomlPath, output)
	versions, err := getNginxVersions()
	if err != nil {
		fmt.Println("error with getNginxVersions")
		os.Exit(0)
	}
	fmt.Printf("\tgetNginxVersions returned: %s\n", versions)

	var versionMetadata []cargo.ConfigMetadataDependency
	for _, version := range versions {
		metadata, err := generateMetadata(version)
		if err != nil {
			fmt.Println("error with generateMetadata")
			os.Exit(0)
		}
		versionMetadata = append(versionMetadata, metadata)
	}
	fmt.Printf("\tgenerateMetadata returned: %s\n", versionMetadata)
}

func getNginxVersions() ([]string, error) {
	// curl https://api.github.com/repos/nginx/nginx/tags
	resp, err := http.Get("https://api.github.com/repos/nginx/nginx/tags")
	if err != nil {
		return nil, fmt.Errorf("could not make request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var tags []GithubTagReponse
	err = json.Unmarshal(body, &tags)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshall tags: %w", err)
	}
	// fmt.Printf("\ttags: %s\n", tags)

	var versions []string
	for _, tag := range tags {
		versions = append(versions, strings.TrimPrefix(tag.Name, "release-"))
	}

	return versions, nil
}

// func generateMetadata
func generateMetadata(version string) (cargo.ConfigMetadataDependency, error) {
	dep := cargo.ConfigMetadataDependency{
		Version: version,
		ID:      "nginx",
		Name:    "NGINX",
		Source:  fmt.Sprintf("http://nginx.org/download/nginx-%s.tar.gz", version),
		// SourceSHA256: "some-sha256", //https://github.com/paketo-buildpacks/dep-server/blob/0b59b7e862517d8b5e68fc6bcee534e4311e1320/pkg/dependency/nginx.go#L74
		// PURL:         "some-purl",
		// CPE:          "some-cpe",
		// Licenses:     "some-license", //https://github.com/paketo-buildpacks/dep-server/blob/0b59b7e862517d8b5e68fc6bcee534e4311e1320/pkg/dependency/licenses/licenses.go#L22
	}

	return dep, nil

}
