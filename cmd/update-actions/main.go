package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	defaultFile   = "renderers/actions/external_actions.go"
	defaultE2EDir = ".e2e"
)

// shaLineRe matches lines like:
//
//	Checkout: ExternalAction{Ref: "actions/checkout@abcdef1234567890abcdef1234567890abcdef12", Version: "v6.0.2"},
//
// It captures: (1) owner/repo, (2) sha, (3) version string (e.g. "v6.0.2").
var shaLineRe = regexp.MustCompile(`Ref:\s*"([^"@]+)@([a-f0-9]{40})",\s*Version:\s*"(v\d+(?:\.\d+)*)"`)

type action struct {
	owner      string
	repo       string
	currentSHA string
	currentVer string // full version string from comment (e.g. "5" or "5.0.1")
	lineIndex  int
}

type release struct {
	TagName string `json:"tag_name"`
}

type tagEntry struct {
	Name string `json:"name"`
}

type gitRef struct {
	Object struct {
		Type string `json:"type"`
		SHA  string `json:"sha"`
		URL  string `json:"url"`
	} `json:"object"`
}

type tagObject struct {
	Object struct {
		Type string `json:"type"`
		SHA  string `json:"sha"`
	} `json:"object"`
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Show what would change without modifying the file")
	flag.Parse()

	filePath := defaultFile

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_TOKEN not set — using unauthenticated requests (rate limit: 60/hour)")
	} else {
		fmt.Println("Using GITHUB_TOKEN for authentication")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fatalf("Failed to read %s: %v\n", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	actions := parseActions(lines)

	fmt.Printf("Parsing %s...\n", filePath)
	fmt.Printf("Found %d SHA-pinned actions\n\n", len(actions))

	if len(actions) == 0 {
		fmt.Println("Nothing to update.")
		return
	}

	// Sort by owner/repo for stable output.
	sort.Slice(actions, func(i, j int) bool {
		ai := actions[i].owner + "/" + actions[i].repo
		aj := actions[j].owner + "/" + actions[j].repo
		return ai < aj
	})

	// shaUpdates maps old SHA -> new SHA for actions that changed.
	shaUpdates := make(map[string]string)
	// shaToVersion maps SHA -> version tag for updating YAML comments
	shaToVersion := make(map[string]string)
	hasErrors := false
	commentUpdates := 0

	for _, a := range actions {
		fullName := a.owner + "/" + a.repo
		latestTag, err := resolveLatestVersion(token, a.owner, a.repo)
		if err != nil {
			fmt.Printf("  %-45s ERROR: %v\n", fullName, err)
			hasErrors = true
			continue
		}

		latestVer := parseMajorVersion(latestTag)
		currentMajor := parseMajorVersion("v" + a.currentVer)
		latestSHA, err := resolveTagSHA(token, a.owner, a.repo, latestTag)
		if err != nil {
			fmt.Printf("  %-45s ERROR resolving SHA for %s: %v\n", fullName, latestTag, err)
			hasErrors = true
			continue
		}

		// Normalise the tag for display (ensure "v" prefix).
		displayTag := latestTag
		if !strings.HasPrefix(displayTag, "v") {
			displayTag = "v" + displayTag
		}

		if latestSHA == a.currentSHA {
			// Normalise the version comment to the full semver tag even when the SHA hasn't changed.
			newLine := replaceLine(lines[a.lineIndex], latestSHA, displayTag)
			if newLine != lines[a.lineIndex] {
				lines[a.lineIndex] = newLine
				commentUpdates++
				fmt.Printf("  %-45s v%-10s (comment updated to %s)\n", fullName, a.currentVer, displayTag)
			} else {
				fmt.Printf("  %-45s v%-10s (up to date)\n", fullName, a.currentVer)
			}
			continue
		}

		prefix := " "
		suffix := ""
		if latestVer > currentMajor {
			prefix = "!"
			suffix = " WARNING: major version bump!"
		}

		fmt.Printf("%s %-45s v%-10s -> %-10s%s\n",
			prefix, fullName, a.currentVer, displayTag, suffix)

		shaUpdates[a.currentSHA] = latestSHA
		shaToVersion[latestSHA] = displayTag

		// Update the line in place.
		lines[a.lineIndex] = replaceLine(lines[a.lineIndex], latestSHA, displayTag)
	}

	fmt.Println()

	if hasErrors {
		fmt.Println("Some actions could not be checked — see errors above.")
	}

	if len(shaUpdates) == 0 && commentUpdates == 0 {
		fmt.Println("All actions are up to date.")
		return
	}

	if len(shaUpdates) == 0 {
		// Only comment updates — no e2e YAML files need changing.
		if *dryRun {
			fmt.Println("Dry run complete. No changes written.")
			return
		}
		output := strings.Join(lines, "\n")
		if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
			fatalf("Failed to write %s: %v\n", filePath, err)
		}
		fmt.Printf("Updated %s\n", filePath)
		return
	}

	// Update e2e expected workflow files.
	e2eFiles, err := filepath.Glob(filepath.Join(defaultE2EDir, "*", "actions.expected.yml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to glob e2e files: %v\n", err)
	}

	type yamlUpdate struct {
		path         string
		content      string
		replacements int
	}
	var yamlUpdates []yamlUpdate

	if len(e2eFiles) > 0 {
		fmt.Println("Updating e2e expected workflow files...")
		for _, f := range e2eFiles {
			yData, err := os.ReadFile(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: failed to read %s: %v\n", f, err)
				continue
			}
			content := string(yData)
			count := 0
			for oldSHA, newSHA := range shaUpdates {
				n := strings.Count(content, oldSHA)
				if n > 0 {
					// Replace SHA and update version comment
					newVersion := shaToVersion[newSHA]
					// Replace "oldSHA" or "oldSHA # vX.Y.Z" with "newSHA # vX.Y.Z"
					shaWithCommentRe := regexp.MustCompile(oldSHA + `(\s*#\s*v\d+(?:\.\d+)*)?`)
					content = shaWithCommentRe.ReplaceAllString(content, newSHA+" # "+newVersion)
					count += n
				}
			}
			if count > 0 {
				noun := "replacements"
				if count == 1 {
					noun = "replacement"
				}
				fmt.Printf("  %-60s (%d %s)\n", f, count, noun)
				yamlUpdates = append(yamlUpdates, yamlUpdate{path: f, content: content, replacements: count})
			}
		}
		if len(yamlUpdates) == 0 {
			fmt.Println("  (no matching SHAs found in e2e files)")
		}
		fmt.Println()
	}

	if *dryRun {
		fmt.Println("Dry run complete. No changes written.")
		return
	}

	// Write the Go file.
	output := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
		fatalf("Failed to write %s: %v\n", filePath, err)
	}
	fmt.Printf("Updated %s\n", filePath)

	// Write the YAML files.
	for _, u := range yamlUpdates {
		if err := os.WriteFile(u.path, []byte(u.content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", u.path, err)
		} else {
			fmt.Printf("Updated %s\n", u.path)
		}
	}
}

// parseActions scans lines for SHA-pinned GitHub Action references.
func parseActions(lines []string) []action {
	var actions []action
	for i, line := range lines {
		m := shaLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		parts := strings.SplitN(m[1], "/", 2)
		if len(parts) != 2 {
			continue
		}
		// m[3] is now the full version with 'v' prefix (e.g., "v6.0.2")
		// Strip the 'v' prefix for consistency with the rest of the code
		ver := strings.TrimPrefix(m[3], "v")
		actions = append(actions, action{
			owner:      parts[0],
			repo:       parts[1],
			currentSHA: m[2],
			currentVer: ver,
			lineIndex:  i,
		})
	}
	return actions
}

// replaceLine replaces the SHA and version in a line.
func replaceLine(line string, newSHA string, newTag string) string {
	// Replace the 40-char hex SHA.
	repl := regexp.MustCompile(`@[a-f0-9]{40}`)
	line = repl.ReplaceAllString(line, "@"+newSHA)

	// Replace the Version field value with the full semver tag.
	// Matches: Version: "v1.2.3" and replaces with Version: "vX.Y.Z"
	verRepl := regexp.MustCompile(`Version:\s*"v\d+(?:\.\d+)*"`)
	line = verRepl.ReplaceAllString(line, `Version: "`+newTag+`"`)

	return line
}

// resolveLatestVersion finds the latest release tag for a GitHub repo.
// Falls back to the highest semver tag if no releases exist.
func resolveLatestVersion(token, owner, repo string) (string, error) {
	tag, err := latestRelease(token, owner, repo)
	if err == nil && tag != "" {
		return tag, nil
	}
	// Fall back to tags.
	return highestTag(token, owner, repo)
}

func latestRelease(token, owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	body, statusCode, err := ghGet(token, url)
	if err != nil {
		return "", err
	}
	if statusCode == 404 {
		return "", fmt.Errorf("no releases found")
	}
	if statusCode != 200 {
		return "", fmt.Errorf("unexpected status %d: %s", statusCode, truncate(string(body), 200))
	}
	var r release
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("decoding release: %w", err)
	}
	if r.TagName == "" {
		return "", fmt.Errorf("release has no tag_name")
	}
	return r.TagName, nil
}

func highestTag(token, owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags?per_page=100", owner, repo)
	body, statusCode, err := ghGet(token, url)
	if err != nil {
		return "", err
	}
	if statusCode != 200 {
		return "", fmt.Errorf("unexpected status %d fetching tags", statusCode)
	}
	var tags []tagEntry
	if err := json.Unmarshal(body, &tags); err != nil {
		return "", fmt.Errorf("decoding tags: %w", err)
	}
	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found")
	}

	// Find the highest semver tag.
	bestTag := ""
	bestMajor := -1
	bestMinor := -1
	bestPatch := -1

	semverRe := regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?$`)
	for _, t := range tags {
		m := semverRe.FindStringSubmatch(t.Name)
		if m == nil {
			continue
		}
		major, _ := strconv.Atoi(m[1])
		minor := 0
		if m[2] != "" {
			minor, _ = strconv.Atoi(m[2])
		}
		patch := 0
		if m[3] != "" {
			patch, _ = strconv.Atoi(m[3])
		}

		if major > bestMajor ||
			(major == bestMajor && minor > bestMinor) ||
			(major == bestMajor && minor == bestMinor && patch > bestPatch) {
			bestMajor = major
			bestMinor = minor
			bestPatch = patch
			bestTag = t.Name
		}
	}

	if bestTag == "" {
		return "", fmt.Errorf("no semver tags found")
	}
	return bestTag, nil
}

// resolveTagSHA resolves a git tag to its underlying commit SHA.
// Handles both lightweight tags (pointing directly to a commit) and
// annotated tags (pointing to a tag object that in turn points to a commit).
func resolveTagSHA(token, owner, repo, tag string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/tags/%s", owner, repo, tag)
	body, statusCode, err := ghGet(token, url)
	if err != nil {
		return "", err
	}
	if statusCode != 200 {
		return "", fmt.Errorf("unexpected status %d resolving ref for tag %s", statusCode, tag)
	}

	var ref gitRef
	if err := json.Unmarshal(body, &ref); err != nil {
		return "", fmt.Errorf("decoding git ref: %w", err)
	}

	// Lightweight tag — object is the commit directly.
	if ref.Object.Type == "commit" {
		return ref.Object.SHA, nil
	}

	// Annotated tag — need to dereference the tag object to get the commit.
	if ref.Object.Type == "tag" {
		return dereferenceTagObject(token, owner, repo, ref.Object.SHA)
	}

	return "", fmt.Errorf("unexpected ref object type: %s", ref.Object.Type)
}

func dereferenceTagObject(token, owner, repo, tagSHA string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/tags/%s", owner, repo, tagSHA)
	body, statusCode, err := ghGet(token, url)
	if err != nil {
		return "", err
	}
	if statusCode != 200 {
		return "", fmt.Errorf("unexpected status %d dereferencing tag object", statusCode)
	}

	var obj tagObject
	if err := json.Unmarshal(body, &obj); err != nil {
		return "", fmt.Errorf("decoding tag object: %w", err)
	}

	if obj.Object.Type == "commit" {
		return obj.Object.SHA, nil
	}

	return "", fmt.Errorf("tag object points to unexpected type: %s", obj.Object.Type)
}

// parseMajorVersion extracts the major version number from a tag like "v3.2.1".
func parseMajorVersion(tag string) int {
	tag = strings.TrimPrefix(tag, "v")
	parts := strings.SplitN(tag, ".", 2)
	v, _ := strconv.Atoi(parts[0])
	return v
}

func ghGet(token, url string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	return body, resp.StatusCode, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
