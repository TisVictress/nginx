package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/libdependency/retrieve"
	"github.com/joshuatcasey/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/vacation"
	"golang.org/x/crypto/openpgp"
)

type GithubTagReponse struct {
	Name string `json:"name"`
}

type NginxMetadata struct {
	SemverVersion *semver.Version
}

func (nginxMetadata NginxMetadata) GetVersion() *semver.Version {
	return nginxMetadata.SemverVersion
}

func main() {
	retrieve.NewMetadata("nginx", getNginxVersions, generateMetadata, "bionic", "jammy")
}

func getNginxVersions() ([]versionology.HasVersion, error) {
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

	var versions []versionology.HasVersion
	for _, tag := range tags {
		versions = append(versions, NginxMetadata{
			semver.MustParse(strings.TrimPrefix(tag.Name, "release-")),
		})
	}

	return versions, nil
}

func generateMetadata(hasVersion versionology.HasVersion) (cargo.ConfigMetadataDependency, error) {
	nginxVersion := hasVersion.GetVersion().String()
	nginxURL := fmt.Sprintf("https://nginx.org/download/nginx-%s.tar.gz", nginxVersion)

	sourceSHA, err := getDependencySHA(nginxVersion)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, fmt.Errorf("could get sha: %w", err)
	}

	// If the dependency is to be compiled, the SHA256 and URI field from the metadata should be omitted in this step.
	dep := cargo.ConfigMetadataDependency{
		Version:         nginxVersion,
		ID:              "nginx",
		Name:            "NGINX",
		Source:          nginxURL,
		SourceSHA256:    sourceSHA,
		DeprecationDate: nil,
		Licenses:        retrieve.LookupLicenses(nginxURL, decompress),
		PURL:            retrieve.GeneratePURL("nginx", nginxVersion, sourceSHA, nginxURL),
		CPE:             fmt.Sprintf("cpe:2.3:a:nginx:nginx:%s:*:*:*:*:*:*:*", nginxVersion),
	}

	return dep, nil

}

// todo: add func to libdep. util function to get as much data from downloading the artifact in the license step or something

func getDependencySHA(version string) (string, error) {
	url := fmt.Sprintf("https://nginx.org/download/nginx-%s.tar.gz", version)

	nginxGPGKeys, err := getGPGKeys()
	if err != nil {
		return "", fmt.Errorf("could not get GPG keys: %w", err)
	}

	dependencySignature, err := getDependencySignature(version)
	if err != nil {
		return "", fmt.Errorf("could not get dependency signature: %w", err)
	}

	dependencyOutputDir, err := os.MkdirTemp("", "nginx")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	dependencyOutputPath := filepath.Join(dependencyOutputDir, filepath.Base(url))

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not make request: %w", err)
	}
	defer resp.Body.Close()

	file, err := os.OpenFile(dependencyOutputPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		file.Close()
		return "", fmt.Errorf("failed to write to file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close file: %w", err)
	}

	err = verifyASC(dependencySignature, dependencyOutputPath, nginxGPGKeys...)
	if err != nil {
		return "", fmt.Errorf("dependency signature verification failed: %w", err)
	}

	return fs.NewChecksumCalculator().Sum(dependencyOutputPath)
}

func getGPGKeys() ([]string, error) {
	var nginxGPGKeys []string
	for _, keyURL := range []string{
		// Key URLs from https://nginx.org/en/pgp_keys.html
		"http://nginx.org/keys/mdounin.key",
		"http://nginx.org/keys/maxim.key",
		"http://nginx.org/keys/sb.key",
		"http://nginx.org/keys/thresh.key",
	} {
		resp, err := http.Get(keyURL)
		if err != nil {
			return []string{}, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)

		nginxGPGKeys = append(nginxGPGKeys, string(body))
	}

	return nginxGPGKeys, nil
}

func getDependencySignature(version string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://nginx.org/download/nginx-%s.tar.gz.asc", version))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	return string(body), nil
}

func verifyASC(asc, path string, pgpKeys ...string) error {
	if len(pgpKeys) == 0 {
		return errors.New("no pgp keys provided")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	for _, pgpKey := range pgpKeys {
		keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(pgpKey))
		if err != nil {
			return fmt.Errorf("could not read armored key ring: %w", err)
		}

		_, err = openpgp.CheckArmoredDetachedSignature(keyring, file, strings.NewReader(asc))
		if err != nil {
			log.Printf("failed to check signature: %s", err.Error())
			continue
		}
		log.Printf("found valid pgp key")
		return nil
	}

	return errors.New("no valid pgp keys provided")
}

func decompress(artifact io.Reader, destination string) error {
	archive := vacation.NewArchive(artifact)

	err := archive.StripComponents(1).Decompress(destination)
	if err != nil {
		return fmt.Errorf("failed to decompress source file: %w", err)
	}

	return nil
}
