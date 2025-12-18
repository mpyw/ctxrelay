// Package metatest validates that test fixtures in testdata/src follow the structure defined in structure.json.
package metatest

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// Structure represents the master data for test patterns.
type Structure struct {
	// Targets is the ordered list of test target directories.
	Targets []string `json:"targets"`
	// Tests maps test function names to their metadata.
	Tests map[string]TestMeta `json:"tests"`
}

// TestMeta represents metadata for a single test pattern.
type TestMeta struct {
	// Title is the short title that must appear in comments.
	Title string `json:"title"`
	// Description is the longer description that must appear in comments.
	Description string `json:"description"`
	// Level indicates which file this test belongs to: "basic", "advanced", or "evil".
	Level string `json:"level"`
	// Targets lists which target directories should have this test.
	Targets []string `json:"targets"`
}

// loadStructure loads the structure.json file.
func loadStructure(t *testing.T) *Structure {
	t.Helper()

	data, err := os.ReadFile("structure.json")
	if err != nil {
		t.Fatalf("failed to read structure.json: %v", err)
	}

	var s Structure
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("failed to parse structure.json: %v", err)
	}

	return &s
}

// targetIndex returns the index of a target in the global targets list.
func (s *Structure) targetIndex(target string) int {
	for i, t := range s.Targets {
		if t == target {
			return i
		}
	}
	return -1
}

// sortTargetsByGlobalOrder sorts targets according to the global targets order.
func (s *Structure) sortTargetsByGlobalOrder(targets []string) []string {
	sorted := make([]string, len(targets))
	copy(sorted, targets)
	slices.SortFunc(sorted, func(a, b string) int {
		return s.targetIndex(a) - s.targetIndex(b)
	})
	return sorted
}

// FuncInfo holds information about a function found in a test file.
type FuncInfo struct {
	Name     string
	Comments string // All comments associated with the function
	FilePath string
	FileName string // Just the filename without directory
}

// parseTestFiles parses all Go files in a target directory and extracts function info.
func parseTestFiles(t *testing.T, targetDir string) map[string]FuncInfo {
	t.Helper()

	funcs := make(map[string]FuncInfo)
	fset := token.NewFileSet()

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		// Directory might not exist, return empty map
		return funcs
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		filePath := filepath.Join(targetDir, entry.Name())
		f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", filePath, err)
		}

		// Create a map of positions to comment groups
		commentMap := ast.NewCommentMap(fset, f, f.Comments)

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			var comments strings.Builder

			// Get associated comments from comment map
			if cgs := commentMap[fn]; cgs != nil {
				for _, cg := range cgs {
					comments.WriteString(cg.Text())
					comments.WriteString("\n")
				}
			}

			// Also check the doc comment directly
			if fn.Doc != nil {
				comments.WriteString(fn.Doc.Text())
			}

			funcs[fn.Name.Name] = FuncInfo{
				Name:     fn.Name.Name,
				Comments: comments.String(),
				FilePath: filePath,
				FileName: entry.Name(),
			}
		}
	}

	return funcs
}

// levelToFileName returns the expected filename for a given level.
func levelToFileName(level, target string) string {
	// Special cases for targets that use a single file
	switch target {
	case "goroutinecreator":
		return "creator.go"
	case "goroutinederive":
		return "goroutinederive.go"
	case "carrier":
		return "carrier.go"
	}
	// Standard level-based naming
	return level + ".go"
}

// TestStructureValidity validates the structure.json itself.
func TestStructureValidity(t *testing.T) {
	s := loadStructure(t)

	validLevels := map[string]bool{"basic": true, "advanced": true, "evil": true}

	// Check that all targets in tests are valid
	for testName, meta := range s.Tests {
		for _, target := range meta.Targets {
			if s.targetIndex(target) == -1 {
				t.Errorf("test %q references unknown target %q", testName, target)
			}
		}

		// Check that level is valid
		if meta.Level != "" && !validLevels[meta.Level] {
			t.Errorf("test %q has invalid level %q (must be basic, advanced, or evil)", testName, meta.Level)
		}
	}

	// Check for duplicate targets in the global list
	seen := make(map[string]bool)
	for _, target := range s.Targets {
		if seen[target] {
			t.Errorf("duplicate target in global targets list: %q", target)
		}
		seen[target] = true
	}
}

// TestFunctionExistence validates that test functions exist in their target directories.
func TestFunctionExistence(t *testing.T) {
	s := loadStructure(t)
	srcDir := filepath.Join("..", "src")

	for testName, meta := range s.Tests {
		for _, target := range meta.Targets {
			t.Run(fmt.Sprintf("%s/%s", target, testName), func(t *testing.T) {
				targetDir := filepath.Join(srcDir, target)
				funcs := parseTestFiles(t, targetDir)

				if _, exists := funcs[testName]; !exists {
					t.Errorf("function %q not found in target %q", testName, target)
				}
			})
		}
	}
}

// TestFunctionInCorrectFile validates that test functions are in the correct file based on their level.
func TestFunctionInCorrectFile(t *testing.T) {
	s := loadStructure(t)
	srcDir := filepath.Join("..", "src")

	for testName, meta := range s.Tests {
		// Skip if no level specified
		if meta.Level == "" {
			continue
		}

		for _, target := range meta.Targets {
			t.Run(fmt.Sprintf("%s/%s", target, testName), func(t *testing.T) {
				targetDir := filepath.Join(srcDir, target)
				funcs := parseTestFiles(t, targetDir)

				fn, exists := funcs[testName]
				if !exists {
					t.Skipf("function %q not found in target %q", testName, target)
					return
				}

				expectedFileName := levelToFileName(meta.Level, target)
				if fn.FileName != expectedFileName {
					t.Errorf("function %q in %s is in %s but should be in %s (level=%s)",
						testName, target, fn.FileName, expectedFileName, meta.Level)
				}
			})
		}
	}
}

// TestTitleInComments validates that the title appears in function comments.
func TestTitleInComments(t *testing.T) {
	s := loadStructure(t)
	srcDir := filepath.Join("..", "src")

	for testName, meta := range s.Tests {
		for _, target := range meta.Targets {
			t.Run(fmt.Sprintf("%s/%s", target, testName), func(t *testing.T) {
				targetDir := filepath.Join(srcDir, target)
				funcs := parseTestFiles(t, targetDir)

				fn, exists := funcs[testName]
				if !exists {
					t.Skipf("function %q not found in target %q", testName, target)
					return
				}

				if !strings.Contains(fn.Comments, meta.Title) {
					t.Errorf("title %q not found in comments for %s in %s\nComments: %s",
						meta.Title, testName, target, fn.Comments)
				}
			})
		}
	}
}

// TestDescriptionInComments validates that the description appears in function comments.
// Note: This test uses Log instead of Error because different targets may have
// target-specific descriptions (e.g., "errgroup.Go()" vs "sync.WaitGroup.Go()").
func TestDescriptionInComments(t *testing.T) {
	s := loadStructure(t)
	srcDir := filepath.Join("..", "src")

	for testName, meta := range s.Tests {
		for _, target := range meta.Targets {
			t.Run(fmt.Sprintf("%s/%s", target, testName), func(t *testing.T) {
				targetDir := filepath.Join(srcDir, target)
				funcs := parseTestFiles(t, targetDir)

				fn, exists := funcs[testName]
				if !exists {
					t.Skipf("function %q not found in target %q", testName, target)
					return
				}

				if !strings.Contains(fn.Comments, meta.Description) {
					// Use Log instead of Error because descriptions may be target-specific
					t.Logf("Note: description %q not found in comments for %s in %s (may have target-specific description)",
						meta.Description, testName, target)
				}
			})
		}
	}
}

// TestSeeAlsoReferences validates that see also references are present and correctly ordered.
func TestSeeAlsoReferences(t *testing.T) {
	s := loadStructure(t)
	srcDir := filepath.Join("..", "src")

	// Regex to extract see also targets
	seeAlsoRegex := regexp.MustCompile(`(?i)see\s+also:\s*(.+)`)

	for testName, meta := range s.Tests {
		// Only check if there are multiple targets
		if len(meta.Targets) <= 1 {
			continue
		}

		for _, target := range meta.Targets {
			t.Run(fmt.Sprintf("%s/%s", target, testName), func(t *testing.T) {
				targetDir := filepath.Join(srcDir, target)
				funcs := parseTestFiles(t, targetDir)

				fn, exists := funcs[testName]
				if !exists {
					t.Skipf("function %q not found in target %q", testName, target)
					return
				}

				// Calculate expected "see also" targets (all targets except current)
				var expectedTargets []string
				for _, tgt := range meta.Targets {
					if tgt != target {
						expectedTargets = append(expectedTargets, tgt)
					}
				}
				expectedTargets = s.sortTargetsByGlobalOrder(expectedTargets)

				// Extract "see also" from comments
				matches := seeAlsoRegex.FindStringSubmatch(fn.Comments)
				if matches == nil {
					t.Errorf("missing 'see also:' in comments for %s in %s (expected: %v)",
						testName, target, expectedTargets)
					return
				}

				// Parse the see also list
				seeAlsoText := strings.TrimSpace(matches[1])
				actualTargets := parseSeeAlsoList(seeAlsoText)

				// Check order and content
				if len(actualTargets) != len(expectedTargets) {
					t.Errorf("see also count mismatch for %s in %s:\nexpected: %v\nactual: %v",
						testName, target, expectedTargets, actualTargets)
					return
				}

				for i := range expectedTargets {
					if actualTargets[i] != expectedTargets[i] {
						t.Errorf("see also mismatch at position %d for %s in %s:\nexpected: %v\nactual: %v",
							i, testName, target, expectedTargets, actualTargets)
						return
					}
				}
			})
		}
	}
}

// parseSeeAlsoList parses a comma-separated list of targets from "see also" text.
func parseSeeAlsoList(text string) []string {
	// Remove trailing newlines and punctuation
	text = strings.TrimRight(text, ".\n\r\t ")

	parts := strings.Split(text, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// TestNoUnregisteredTests checks that all test functions in target directories
// that follow naming conventions are registered in structure.json.
func TestNoUnregisteredTests(t *testing.T) {
	s := loadStructure(t)
	srcDir := filepath.Join("..", "src")

	// Patterns that indicate a test function
	testPrefixes := []string{"bad", "good"}

	for _, target := range s.Targets {
		t.Run(target, func(t *testing.T) {
			targetDir := filepath.Join(srcDir, target)
			funcs := parseTestFiles(t, targetDir)

			for funcName := range funcs {
				// Check if function name starts with test prefix
				isTestFunc := false
				for _, prefix := range testPrefixes {
					if strings.HasPrefix(strings.ToLower(funcName), prefix) {
						isTestFunc = true
						break
					}
				}

				if !isTestFunc {
					continue
				}

				// Check if this function is registered
				meta, registered := s.Tests[funcName]
				if !registered {
					// Not in structure.json at all - might be intentional
					// Only warn, don't fail (for gradual migration)
					t.Logf("WARN: function %q in %s is not registered in structure.json", funcName, target)
					continue
				}

				// Check if this target is in the test's targets list
				found := false
				for _, tgt := range meta.Targets {
					if tgt == target {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("function %q exists in %s but target not listed in structure.json targets: %v",
						funcName, target, meta.Targets)
				}
			}
		})
	}
}
