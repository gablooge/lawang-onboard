// Package roles parses roles.yaml and provides the Role type used throughout the server.
package roles

import "strings"

// matchGlob reports whether the slash-separated path matches the slash-separated pattern.
// The pattern may contain:
//   - "**" matching zero or more slash-separated segments (any run, including empty)
//   - "*"  matching exactly one segment that contains no slash
//   - any other segment matching that literal segment only
//
// Matching is case-sensitive. The empty string matches the empty string.
func matchGlob(pattern, path string) bool {
	return matchSegments(splitSegments(pattern), splitSegments(path))
}

// splitSegments splits a slash-separated path into its segments.
// An empty string returns a nil slice (no segments).
func splitSegments(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "/")
}

// matchSegments is the recursive core of matchGlob.
func matchSegments(patSegs, pathSegs []string) bool {
	for len(patSegs) > 0 {
		seg := patSegs[0]
		patSegs = patSegs[1:]

		if seg == "**" {
			// "**" matches zero or more path segments. Try every possible length.
			// Zero segments: consume "**" and continue matching.
			if matchSegments(patSegs, pathSegs) {
				return true
			}
			// One or more segments: consume one path segment and try again with "**" still present.
			for len(pathSegs) > 0 {
				pathSegs = pathSegs[1:]
				if matchSegments(patSegs, pathSegs) {
					return true
				}
			}
			return false
		}

		// No more path segments but still have non-** pattern.
		if len(pathSegs) == 0 {
			return false
		}

		// "*" matches exactly one segment (no slash; since we split on slash, each segment has none).
		if seg == "*" {
			pathSegs = pathSegs[1:]
			continue
		}

		// Literal match.
		if seg != pathSegs[0] {
			return false
		}
		pathSegs = pathSegs[1:]
	}

	return len(pathSegs) == 0
}
