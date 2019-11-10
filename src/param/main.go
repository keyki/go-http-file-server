package param

import (
	"../util"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

type user struct {
	Username string
	Password string
}

type Param struct {
	Root    string
	Aliases map[string]string

	GlobalUpload bool
	UploadUrls   []string
	UploadDirs   []string

	GlobalArchive bool
	ArchiveUrls   []string
	ArchiveDirs   []string

	GlobalCors bool
	CorsUrls   []string
	CorsDirs   []string

	GlobalAuth  bool
	AuthUrls    []string
	AuthDirs    []string
	UsersPlain  []*user
	UsersBase64 []*user
	UsersMd5    []*user
	UsersSha1   []*user
	UsersSha256 []*user
	UsersSha512 []*user

	Key         string
	Cert        string
	Listen      []string
	ListenPlain []string
	ListenTLS   []string
	Hostnames   []string
	Template    string

	Shows     *regexp.Regexp
	ShowDirs  *regexp.Regexp
	ShowFiles *regexp.Regexp
	Hides     *regexp.Regexp
	HideDirs  *regexp.Regexp
	HideFiles *regexp.Regexp

	AccessLog string
	ErrorLog  string
}

func normalizePathMaps(inputs []string) map[string]string {
	maps := map[string]string{}

	for _, input := range inputs {
		sep, sepLen := utf8.DecodeRuneInString(input)
		if sepLen == 0 {
			continue
		}
		input = input[sepLen:]
		if len(input) == 0 {
			continue
		}

		sepIndex := strings.Index(input, string(sep))
		if sepIndex == -1 {
			continue
		}

		urlPath := input[:sepIndex]
		if len(urlPath) == 0 {
			continue
		}
		urlPath = util.CleanUrlPath(urlPath)

		fsPath := input[sepIndex+sepLen:]
		if len(fsPath) == 0 {
			continue
		}
		fsPath = path.Clean(fsPath)

		maps[urlPath] = fsPath
	}

	return maps
}

func normalizeUrlPaths(inputs []string) []string {
	outputs := make([]string, 0, len(inputs))

	for _, input := range inputs {
		if len(input) == 0 {
			continue
		}
		outputs = append(outputs, util.CleanUrlPath(input))
	}

	return outputs
}

func normalizeFsPaths(inputs []string) []string {
	outputs := make([]string, 0, len(inputs))

	for _, input := range inputs {
		if len(input) == 0 {
			continue
		}

		abs, err := filepath.Abs(input)
		if err != nil {
			continue
		}

		outputs = append(outputs, abs)
	}

	return outputs
}

func getWildcardRegexp(wildcards []string, found bool) (*regexp.Regexp, error) {
	if !found || len(wildcards) == 0 {
		return nil, nil
	}

	normalizedWildcards := make([]string, 0, len(wildcards))
	for _, wildcard := range wildcards {
		if len(wildcard) == 0 {
			continue
		}
		normalizedWildcards = append(normalizedWildcards, util.WildcardToRegexp(wildcard))
	}

	if len(normalizedWildcards) == 0 {
		return nil, nil
	}

	exp := strings.Join(normalizedWildcards, "|")
	return regexp.Compile(exp)
}
